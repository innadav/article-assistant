# Refactored Code Tests

This directory contains test scripts for the refactored article assistant implementation.

## Test Scripts

### `test_refactored_concepts.sh`
Comprehensive test suite that validates the two-phase LLM approach concepts:
- **Phase 1 (Planning)**: Intent recognition and parameter extraction
- **Phase 2 (Execution)**: Response generation and content analysis
- **Traceability**: Source attribution validation

**Usage:**
```bash
./test_refactored_concepts.sh
```

**Tests Covered:**
1. Two-Phase Summary Test
2. Two-Phase Keywords Test
3. Two-Phase Sentiment Test
4. Two-Phase Topic Search Test
5. Two-Phase Entity Analysis Test
6. Two-Phase Comparison Test
7. Two-Phase Tone Analysis Test
8. Two-Phase Positive Sentiment Test

### `test_summary_demo.sh`
Quick demo script that demonstrates the two-phase LLM approach with a single summary query.

**Usage:**
```bash
./test_summary_demo.sh
```

## Refactored Code Location

The actual refactored implementation code is located in:
```
/Users/innad/workspace/article-assistant/refactored/
```

## Running Tests

**Prerequisites:**
- Refactored server must be running (when build issues are resolved)
- Or tests will run against production server for concept validation

**To run all tests:**
```bash
cd /Users/innad/workspace/article-assistant/tests/refactored
./test_refactored_concepts.sh
```

**To run quick demo:**
```bash
./test_summary_demo.sh
```

## Test Results

Test results are saved to:
```
/Users/innad/workspace/article-assistant/results/refactored_test/
├── results_<timestamp>.json  # Detailed JSON results
└── report_<timestamp>.txt    # Human-readable report
```

## Notes

- Currently tests validate concepts using the production server
- Once refactored code builds, update scripts to point to refactored server
- All tests include two-phase analysis (planning, execution, traceability)
