import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "skills/openydt-record/scripts/parking_duration.py"


class ParkingDurationScriptTest(unittest.TestCase):
    def run_script(self, records, *arguments):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        input_path = Path(directory.name) / "records.ndjson"
        input_path.write_text(
            "".join(json.dumps(record) + "\n" for record in records),
            encoding="utf-8",
        )
        result = subprocess.run(
            [sys.executable, str(SCRIPT), str(input_path), *arguments],
            text=True,
            capture_output=True,
            check=False,
        )
        return result, Path(directory.name)

    def test_missing_enter_time_is_excluded_from_duration_only(self):
        records = [
            {
                "parkingCode": "valid",
                "enterTime": "20260801080000",
                "leaveTime": "20260801100000",
                "leaveType": 1,
            },
            {
                "parkingCode": "missing-enter",
                "leaveTime": "20260801110000",
                "stoppingTime": 120,
                "leaveType": 1,
            },
        ]

        with tempfile.TemporaryDirectory() as directory:
            input_path = Path(directory) / "records.ndjson"
            anomaly_path = Path(directory) / "anomalies.ndjson"
            input_path.write_text(
                "".join(json.dumps(record) + "\n" for record in records),
                encoding="utf-8",
            )

            result = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    str(input_path),
                    "--anomalies-out",
                    str(anomaly_path),
                ],
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["operatingDepartureRecords"], 2)
            self.assertEqual(summary["validDurationRecords"], 1)
            self.assertEqual(summary["excludedDurationRecords"], 1)
            self.assertEqual(summary["missingEnterTime"], 1)
            self.assertEqual(summary["durationCoverageRate"], 0.5)
            self.assertEqual(summary["inputFiles"], [str(input_path)])
            self.assertEqual(summary["durationHours"]["mean"], 2.0)

            anomalies = [
                json.loads(line)
                for line in anomaly_path.read_text(encoding="utf-8").splitlines()
            ]
            self.assertEqual(len(anomalies), 1)
            self.assertEqual(anomalies[0]["reason"], "missing_enter_time")
            self.assertEqual(anomalies[0]["record"]["parkingCode"], "missing-enter")

    def test_reports_default_distribution_and_nearest_rank_percentiles(self):
        records = []
        for index, hours in enumerate((0.5, 1, 4, 12, 13)):
            whole_hours = int(hours)
            minutes = int((hours - whole_hours) * 60)
            records.append(
                {
                    "parkingCode": str(index),
                    "enterTime": "20260801000000",
                    "leaveTime": f"20260801{whole_hours:02d}{minutes:02d}00",
                    "leaveType": 1,
                }
            )

        result, _ = self.run_script(records)

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["durationHours"]["p50"], 4.0)
        self.assertEqual(summary["durationHours"]["p90"], 13.0)
        self.assertEqual(
            [bucket["count"] for bucket in summary["durationDistribution"]],
            [1, 1, 1, 0, 0, 2],
        )

    def test_anomaly_output_is_private_even_when_file_already_exists(self):
        records = [
            {"parkingCode": "missing", "leaveTime": "20260801110000", "leaveType": 1}
        ]
        result, directory = self.run_script(records)
        self.assertEqual(result.returncode, 0, result.stderr)

        anomaly_path = directory / "anomalies.ndjson"
        anomaly_path.write_text("old\n", encoding="utf-8")
        anomaly_path.chmod(0o644)
        result, _ = self.run_script(
            records, "--anomalies-out", str(anomaly_path)
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(os.stat(anomaly_path).st_mode & 0o777, 0o600)

    def test_classifies_other_time_failures_and_nested_platform_records(self):
        records = [
            {
                "parkingCode": "nested-valid",
                "data": {
                    "enterTime": "2026-08-01 08:00:00",
                    "leaveTime": "2026-08-01 09:00:00",
                    "leaveType": 1,
                },
            },
            {
                "parkingCode": "missing-leave",
                "enterTime": "20260801080000",
                "leaveType": 1,
            },
            {
                "parkingCode": "invalid-enter",
                "enterTime": "not-a-time",
                "leaveTime": "20260801090000",
                "leaveType": 1,
            },
            {
                "parkingCode": "invalid-leave",
                "enterTime": "20260801080000",
                "leaveTime": "not-a-time",
                "leaveType": 1,
            },
            {
                "parkingCode": "negative",
                "enterTime": "20260801100000",
                "leaveTime": "20260801090000",
                "leaveType": 1,
            },
        ]

        result, _ = self.run_script(records)

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["operatingDepartureRecords"], 5)
        self.assertEqual(summary["validDurationRecords"], 1)
        self.assertEqual(summary["missingLeaveTime"], 1)
        self.assertEqual(summary["invalidEnterTime"], 1)
        self.assertEqual(summary["invalidLeaveTime"], 1)
        self.assertEqual(summary["negativeDuration"], 1)

    def test_deduplicates_by_parking_code_and_keeps_latest_departure(self):
        records = [
            {
                "parkingCode": "same-session",
                "enterTime": "20260801080000",
                "leaveTime": "20260801100000",
                "leaveType": 1,
            },
            {
                "parkingCode": "same-session",
                "enterTime": "20260801080000",
                "leaveTime": "20260801100500",
                "leaveType": 12,
            },
        ]

        result, _ = self.run_script(records)

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["rawDepartureEvents"], 2)
        self.assertEqual(summary["deduplicatedDepartureEvents"], 1)
        self.assertEqual(summary["duplicateDepartureEvents"], 1)
        self.assertEqual(summary["operatingDepartureRecords"], 1)
        self.assertEqual(summary["validDurationRecords"], 1)
        self.assertAlmostEqual(summary["durationHours"]["mean"], 2.083333, places=6)
        self.assertEqual(summary["leaveTypeDistribution"], [
            {"code": 12, "count": 1, "businessTags": ["repeat_departure"]}
        ])

    def test_excludes_non_physical_closures_and_counts_tailgating_as_escape(self):
        records = [
            {
                "parkingCode": "normal",
                "enterTime": "20260801080000",
                "leaveTime": "20260801090000",
                "leaveType": 1,
            },
            {
                "parkingCode": "escape",
                "enterTime": "20260801080000",
                "leaveTime": "20260801100000",
                "leaveType": 6,
            },
            {
                "parkingCode": "legacy",
                "enterTime": "20260801080000",
                "leaveTime": "20260801110000",
                "leaveType": 13,
            },
            {
                "parkingCode": "turn-back",
                "enterTime": "20260801080000",
                "leaveTime": "20260801120000",
                "leaveType": 18,
            },
            {
                "parkingCode": "logical-close",
                "enterTime": "20250801080000",
                "leaveTime": "20260801130000",
                "leaveType": 97,
            },
            {
                "parkingCode": "inventory",
                "enterTime": "20250801080000",
                "leaveTime": "20260801140000",
                "leaveType": 98,
            },
        ]

        result, _ = self.run_script(records)

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["rawDepartureEvents"], 6)
        self.assertEqual(summary["deduplicatedDepartureEvents"], 6)
        self.assertEqual(summary["operatingDepartureRecords"], 2)
        self.assertEqual(summary["excludedBusinessRecords"], 4)
        self.assertEqual(summary["validDurationRecords"], 2)
        self.assertEqual(summary["escapeRecords"], 1)
        self.assertEqual(summary["escapeDefinition"], "leaveType=6（可疑跟车）")
        self.assertEqual(
            summary["businessExclusions"],
            [
                {"code": 13, "count": 1, "eventNature": "legacy"},
                {"code": 18, "count": 1, "eventNature": "logical_departure"},
                {"code": 97, "count": 1, "eventNature": "logical_closure"},
                {"code": 98, "count": 1, "eventNature": "manual_reconciliation"},
            ],
        )


if __name__ == "__main__":
    unittest.main()
