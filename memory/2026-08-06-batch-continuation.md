# Batch Continuation Debug Report

- Symptom: Batch continuation failed when every visible household was selected.
- Root cause: The room selector included the `总表` master-meter record. The UI's household statistics excluded it, but the batch request did not. Batch creation is atomic, so an invalid master-meter continuation caused the whole request to roll back.
- Fix: Exclude `总表` from the selector, reject it in both continuation request DTOs, and expose the API's detailed error message in the notification.
- Evidence: `TestBatchContinueWorksWithSQLite` passes against a temporary SQLite database. `go test ./...`, `go vet ./...`, JavaScript syntax checks, and `git diff --check` all passed.
- JSON audit: HTTP JSON request bodies remain appropriate. The report `URLSearchParams` serialization issue was fixed earlier in commit `90841ae`; no other SQLite-related JSON serialization issue was found.
- Status: DONE.
