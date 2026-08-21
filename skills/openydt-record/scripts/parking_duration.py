#!/usr/bin/env python3
"""Build the operating-departure cohort and analyze parking durations."""

import argparse
import json
import math
import os
import statistics
import subprocess
import sys
from collections import Counter
from datetime import datetime
from pathlib import Path


TIME_FORMATS = ("%Y%m%d%H%M%S", "%Y-%m-%d %H:%M:%S", "%Y-%m-%dT%H:%M:%S")
DEFAULT_BUCKETS = (
    ("0-1", 0, 1),
    ("1-4", 1, 4),
    ("4-7", 4, 7),
    ("7-10", 7, 10),
    ("10-12", 10, 12),
    ("12+", 12, None),
)


def parse_args():
    parser = argparse.ArgumentParser(
        description="按领域语义形成经营离场样本，再根据 leaveTime-enterTime 计算停车时长"
    )
    parser.add_argument("inputs", nargs="+", help="一个或多个 NDJSON 文件")
    parser.add_argument("--anomalies-out", help="异常记录 NDJSON 输出文件")
    parser.add_argument(
        "--semantics-file",
        help="领域语义 JSON；默认读取仓库 catalog 或调用 openydt schema",
    )
    return parser.parse_args()


def parse_time(value):
    if not isinstance(value, str) or not value.strip():
        return None
    text = value.strip()
    for time_format in TIME_FORMATS:
        try:
            return datetime.strptime(text, time_format)
        except ValueError:
            pass
    return None


def record_data(record):
    nested = record.get("data")
    if isinstance(nested, dict) and (
        "enterTime" in nested or "leaveTime" in nested
    ):
        return nested
    return record


def read_records(paths):
    for path_text in paths:
        path = Path(path_text)
        with path.open(encoding="utf-8") as source:
            for line_number, line in enumerate(source, start=1):
                if not line.strip():
                    continue
                try:
                    record = json.loads(line)
                except json.JSONDecodeError as error:
                    raise ValueError(f"{path}:{line_number}: 非法 JSON: {error.msg}") from error
                if not isinstance(record, dict):
                    raise ValueError(f"{path}:{line_number}: 每行必须是 JSON 对象")
                yield str(path), line_number, record


def load_semantics(path_text=None):
    if path_text:
        with Path(path_text).open(encoding="utf-8") as source:
            payload = json.load(source)
    else:
        repository_source = (
            Path(__file__).resolve().parents[3] / "catalog" / "domain-semantics.json"
        )
        if repository_source.exists():
            with repository_source.open(encoding="utf-8") as source:
                payload = json.load(source)
        else:
            completed = subprocess.run(
                ["openydt", "schema", "getCarOutList", "--json"],
                text=True,
                capture_output=True,
                check=False,
            )
            if completed.returncode != 0:
                raise ValueError(
                    "无法读取停车领域语义；请升级 openydt 或传 --semantics-file: "
                    + completed.stderr.strip()
                )
            payload = json.loads(completed.stdout)

    if "interfaces" in payload:
        semantics = payload["interfaces"].get("getCarOutList")
    else:
        semantics = payload.get("domainSemantics")
    if not isinstance(semantics, dict):
        raise ValueError("领域语义缺少 getCarOutList")
    return semantics


def field_value(original, record, name):
    if name in record:
        return record.get(name)
    return original.get(name)


def normalize_code(value):
    if value is None or value == "":
        return None
    try:
        return int(value)
    except (TypeError, ValueError):
        return str(value)


def code_sort_key(value):
    if isinstance(value, int):
        return 0, value
    return 1, str(value)


def leave_type_semantics(semantics):
    for field in semantics.get("fields", []):
        if field.get("field") == "leaveType":
            return {value["code"]: value for value in field.get("values", [])}
    raise ValueError("领域语义缺少 getCarOutList.leaveType")


def departure_order(entry):
    original = entry[2]
    record = record_data(original)
    parsed = parse_time(field_value(original, record, "leaveTime"))
    return parsed is not None, parsed or datetime.min, entry[3]


def deduplicate(records, semantics):
    rule = semantics.get("deduplication") or {}
    if rule != {"key": "parkingCode", "orderBy": "leaveTime", "keep": "latest"}:
        raise ValueError("不支持的离场去重规则")

    selected = {}
    missing_key = 0
    for sequence, (source, line_number, original) in enumerate(records):
        record = record_data(original)
        parking_code = field_value(original, record, "parkingCode")
        if parking_code is None or parking_code == "":
            missing_key += 1
            key = ("__row__", source, line_number)
        else:
            key = ("parkingCode", str(parking_code))
        candidate = (source, line_number, original, sequence)
        previous = selected.get(key)
        if previous is None or departure_order(candidate) >= departure_order(previous):
            selected[key] = candidate
    ordered = sorted(selected.values(), key=lambda entry: entry[3])
    return ordered, missing_key


def nearest_rank(values, percentile):
    ordered = sorted(values)
    index = max(0, math.ceil(percentile * len(ordered)) - 1)
    return ordered[index]


def duration_distribution(durations):
    distribution = []
    for label, lower, upper in DEFAULT_BUCKETS:
        count = sum(
            1
            for duration in durations
            if duration >= lower and (upper is None or duration < upper)
        )
        distribution.append(
            {
                "label": label,
                "count": count,
                "shareOfValid": round(count / len(durations), 6)
                if durations
                else 0,
            }
        )
    return distribution


def analyze(paths, semantics):
    counts = Counter()
    durations = []
    anomalies = []
    raw_records = list(read_records(paths))
    records, missing_parking_code = deduplicate(raw_records, semantics)
    policies = leave_type_semantics(semantics)
    type_counts = Counter()
    type_tags = {}
    business_exclusions = Counter()

    for source, line_number, original, _ in records:
        record = record_data(original)
        leave_type = normalize_code(field_value(original, record, "leaveType"))
        policy = policies.get(leave_type)
        if policy is None:
            policy = {
                "eventNature": "unclassified",
                "trafficTreatment": "exclude",
                "durationTreatment": "exclude",
                "businessTags": [],
            }
            counts["unclassified_leave_type"] += 1

        if policy.get("trafficTreatment") == "include":
            counts["total_departures"] += 1
            type_counts[leave_type] += 1
            type_tags[leave_type] = policy.get("businessTags", [])
            if "escape" in policy.get("businessTags", []):
                counts["escape_records"] += 1
        else:
            business_exclusions[(leave_type, policy.get("eventNature", "unclassified"))] += 1

        if policy.get("durationTreatment") != "include":
            continue

        counts["duration_candidates"] += 1
        enter_value = field_value(original, record, "enterTime")
        leave_value = field_value(original, record, "leaveTime")

        reason = None
        if enter_value is None or enter_value == "":
            reason = "missing_enter_time"
            counts[reason] += 1
        elif leave_value is None or leave_value == "":
            reason = "missing_leave_time"
            counts[reason] += 1
        else:
            enter_time = parse_time(enter_value)
            leave_time = parse_time(leave_value)
            if enter_time is None:
                reason = "invalid_enter_time"
                counts[reason] += 1
            elif leave_time is None:
                reason = "invalid_leave_time"
                counts[reason] += 1
            else:
                duration_hours = (leave_time - enter_time).total_seconds() / 3600
                if duration_hours < 0:
                    reason = "negative_duration"
                    counts[reason] += 1
                else:
                    durations.append(duration_hours)

        if reason:
            anomalies.append(
                {
                    "reason": reason,
                    "source": source,
                    "line": line_number,
                    "record": original,
                }
            )

    total = counts["total_departures"]
    valid = len(durations)
    duration_summary = None
    if durations:
        duration_summary = {
            "min": round(min(durations), 6),
            "max": round(max(durations), 6),
            "mean": round(statistics.fmean(durations), 6),
            "median": round(statistics.median(durations), 6),
            "p50": round(nearest_rank(durations, 0.50), 6),
            "p75": round(nearest_rank(durations, 0.75), 6),
            "p90": round(nearest_rank(durations, 0.90), 6),
            "p95": round(nearest_rank(durations, 0.95), 6),
        }

    summary = {
        "inputFiles": [str(Path(path)) for path in paths],
        "rawDepartureEvents": len(raw_records),
        "deduplicatedDepartureEvents": len(records),
        "duplicateDepartureEvents": len(raw_records) - len(records),
        "missingParkingCode": missing_parking_code,
        "operatingDepartureRecords": total,
        "excludedBusinessRecords": len(records) - total,
        "unclassifiedLeaveTypeRecords": counts["unclassified_leave_type"],
        "businessExclusions": [
            {"code": code, "count": count, "eventNature": nature}
            for (code, nature), count in sorted(
                business_exclusions.items(),
                key=lambda item: (code_sort_key(item[0][0]), item[0][1]),
            )
        ],
        "leaveTypeDistribution": [
            {
                "code": code,
                "count": count,
                "businessTags": type_tags.get(code, []),
            }
            for code, count in sorted(type_counts.items(), key=lambda item: code_sort_key(item[0]))
        ],
        "escapeRecords": counts["escape_records"],
        "escapeDefinition": "leaveType=6（可疑跟车）",
        "escapeShareOfOperatingDepartures": round(
            counts["escape_records"] / total, 6
        )
        if total
        else 0,
        "durationCandidateRecords": counts["duration_candidates"],
        "validDurationRecords": valid,
        "excludedDurationRecords": counts["duration_candidates"] - valid,
        "missingEnterTime": counts["missing_enter_time"],
        "missingLeaveTime": counts["missing_leave_time"],
        "invalidEnterTime": counts["invalid_enter_time"],
        "invalidLeaveTime": counts["invalid_leave_time"],
        "negativeDuration": counts["negative_duration"],
        "invalidTime": counts["invalid_enter_time"]
        + counts["invalid_leave_time"]
        + counts["negative_duration"],
        "durationCoverageRate": round(valid / counts["duration_candidates"], 6)
        if counts["duration_candidates"]
        else 0,
        "durationHours": duration_summary,
        "durationDistribution": duration_distribution(durations),
    }
    return summary, anomalies


def write_anomalies(path_text, anomalies):
    if not path_text:
        return
    descriptor = os.open(path_text, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    if hasattr(os, "fchmod"):
        os.fchmod(descriptor, 0o600)
    else:
        os.chmod(path_text, 0o600)
    with os.fdopen(descriptor, "w", encoding="utf-8") as output:
        for anomaly in anomalies:
            output.write(json.dumps(anomaly, ensure_ascii=False, separators=(",", ":")))
            output.write("\n")


def main():
    args = parse_args()
    try:
        semantics = load_semantics(args.semantics_file)
        summary, anomalies = analyze(args.inputs, semantics)
        write_anomalies(args.anomalies_out, anomalies)
    except (OSError, ValueError) as error:
        print(f"parking-duration: {error}", file=sys.stderr)
        return 2
    json.dump(summary, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
