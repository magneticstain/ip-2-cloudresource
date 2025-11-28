<!-- markdownlint-disable MD013 -->
# GitHub Issues Templates

Copy and paste each section into a new GitHub Issue. All issues are for the `ip-2-cloudresource` repository.

---

## PRIORITY 2 (HIGH) - Do Soon

### Issue 1: Fix Race Condition in Goroutine Workers

**Title:** Fix race condition in search worker goroutines

**Description:**

There is a concurrent modification of shared state without proper synchronization in the search worker goroutines.

**Location:** `search/search.go:160-168` in the `runSearchWorker` function

**Current Problem:**

```go
if acctID != "current" && search.Platform == "aws" {
    // ...
    search.AWSCtrlr.PrincipalAWSConn = ac  // ⚠️ Unsafe concurrent modification
}
```

Multiple goroutines are modifying `search.AWSCtrlr.PrincipalAWSConn` simultaneously without synchronization, which can cause data races and unpredictable behavior during concurrent account searches.

**Suggested Solutions:**

1. Use immutable copies of the connector per goroutine
2. Protect the shared state with `sync.Mutex`
3. Pass credentials as function parameters instead of modifying receiver state

**Acceptance Criteria:**

- [ ] No concurrent modification of shared AWS connector state
- [ ] Unit tests verify thread-safe behavior under concurrent load
- [ ] `go run -race` reports no data races
- [ ] All existing tests pass

**Labels:** `bug`, `concurrency`, `high-priority`

---

### Issue 2: Replace log.Fatal() with Proper Error Returns

**Title:** Replace log.Fatal() calls with proper error returns

**Description:**

The application currently uses `log.Fatal()` in recoverable error scenarios, which causes abrupt application termination instead of graceful error handling.

**Locations:**

1. `search/search.go:214` - error connecting to platform
2. `app/app.go:86` - unsupported platform validation
3. `app/app.go:110` - general error in cloud search

**Current Pattern:**

```go
log.Fatal("error when connecting to ", search.Platform, ": ", err)
```

**Better Pattern:**

```go
return false, fmt.Errorf("error connecting to %s: %w", search.Platform, err)
```

**Why This Matters:**

- Enables proper error handling at the call site
- Allows applications embedding this library to handle errors gracefully
- Improves testability of error scenarios

**Acceptance Criteria:**

- [ ] All 3 `log.Fatal()` calls replaced with error returns
- [ ] Callers properly handle returned errors
- [ ] Unit tests verify error propagation
- [ ] All existing tests pass

**Labels:** `improvement`, `error-handling`, `high-priority`

---

### Issue 3: Add Context Timeouts to API Calls

**Title:** Add context timeouts to cloud provider API calls

**Description:**

Multiple API calls use `context.TODO()` without timeout protection, which can cause the application to hang indefinitely if cloud provider APIs are unresponsive.

**Affected Files:**

- `aws/plugin/ec2/ec2.go:27`
- `aws/plugin/cloudfront/cloudfront.go:50`
- `aws/plugin/elb/elb_classic.go:29`
- `aws/plugin/elb/elb.go:32, 56, 96`
- `aws/plugin/iam/iam.go:20`
- `aws/plugin/organizations/organizations.go:25, 44`

**Current Pattern:**

```go
output, err := paginator.NextPage(context.TODO())
```

**Desired Pattern:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
output, err := paginator.NextPage(ctx)
```

**Benefits:**

- Prevents indefinite hangs during API calls
- Enables cancellation support
- Improves application responsiveness

**Acceptance Criteria:**

- [ ] All identified API calls have context timeouts
- [ ] Timeouts are configurable or use reasonable defaults (e.g., 30-60 seconds)
- [ ] Unit tests verify timeout behavior
- [ ] All existing tests pass

**Labels:** `improvement`, `reliability`, `high-priority`

---

### Issue 4: Improve Error Handling in Goroutine Workers

**Title:** Improve error handling in account search worker goroutines

**Description:**

Errors occurring in goroutine workers are only logged but not propagated, causing search failures to go unnoticed. Users don't know if a search failed in a particular account.

**Location:** `search/search.go:165-169`

**Current Pattern:**

```go
resultResource, err := search.doAccountLevelSearch(acctID, doNetMapping)
if err != nil {
    log.Error("error when running search within account search worker: ", err)
} else if resultResource.RID != "" {
    matchingResourceBuffer <- resultResource
    return
}
```

**Problem:** Errors are logged but not communicated back to the caller, making it impossible to know which accounts had search failures.

**Suggested Improvements:**

1. Send error information through a dedicated error channel
2. Use `errgroup.Group` for cleaner error collection
3. Return partial results with error information

**Acceptance Criteria:**

- [ ] Worker errors are propagated to caller
- [ ] Failed accounts are clearly communicated
- [ ] Users can distinguish between "not found" and "search error"
- [ ] Unit tests verify error propagation
- [ ] All existing tests pass

**Labels:** `improvement`, `error-handling`, `high-priority`

---

## PRIORITY 3 (MEDIUM) - Next Sprint

### Issue 5: Add Input Validation for Org Search Parameters

**Title:** Add input validation for AWS Organizations search parameters

**Description:**

When org search is enabled, certain parameters should be validated to prevent invalid configurations or confusing behavior.

**Parameters Needing Validation:**

- `org-search-role-name`: Should be validated as a valid IAM role name format
- `org-search-xaccount-role-arn`: Should be validated as a valid ARN format
- `org-search-ou-id`: Should be validated as a valid OU ID format

**Current State:** No validation; invalid inputs silently fail

**Suggested Approach:**

1. Create validation functions in `utils/validation.go`
2. Add validation in `cmd/root.go` during flag parsing
3. Provide clear error messages for invalid inputs

**Acceptance Criteria:**

- [ ] All org search parameters are validated
- [ ] Clear error messages for invalid formats
- [ ] Unit tests for all validation functions
- [ ] Integration tests verify validation during CLI usage
- [ ] All existing tests pass

**Labels:** `improvement`, `validation`, `ux`

---

### Issue 6: Fix Error Return in IP Fuzzing Function

**Title:** Fix error return in IP fuzzing FQDN mapping

**Description:**

In `aws/svc/ip_fuzzing/ip_fuzzing.go`, the `RunAdvancedFuzzing` function catches an error but returns `nil` instead of the actual error.

**Location:** `aws/svc/ip_fuzzing/ip_fuzzing.go:43-45`

**Current Code:**

```go
svcName, err = MapFQDNToSvc(fqdn)
if err != nil {
    return cloudSvc, nil  // ⚠️ Returns nil error despite having one
}
```

**Should Be:**

```go
svcName, err = MapFQDNToSvc(fqdn)
if err != nil {
    return cloudSvc, err  // Return the actual error
}
```

**Impact:** Errors in FQDN-to-service mapping are silently swallowed, making debugging difficult.

**Acceptance Criteria:**

- [ ] Function returns actual errors
- [ ] Error propagates to caller
- [ ] Unit tests verify error is returned
- [ ] All existing tests pass

**Labels:** `bug`, `error-handling`, `medium-priority`

---

## PRIORITY 4 (LOW) - Future Improvements

### Issue 7: Use strings.Builder for Efficient String Concatenation

**Title:** Refactor to use strings.Builder for network map formatting

**Description:**

The network map formatting in `app/app.go` uses string concatenation in a loop, which is less efficient than using `strings.Builder`.

**Location:** `app/app.go:41-48`

**Current Pattern:**

```go
networkMapGraph := ""
for i, networkResource := range matchedResource.NetworkMap {
    networkResourceElmnt = "%s"
    if i != networkMapResourceCnt-1 {
        networkResourceElmnt += " -> "
    }
    networkMapGraph += fmt.Sprintf(networkResourceElmnt, networkResource)
}
```

**Better Pattern:**

```go
var buf strings.Builder
for i, networkResource := range matchedResource.NetworkMap {
    if i > 0 {
        buf.WriteString(" -> ")
    }
    buf.WriteString(networkResource)
}
networkMapGraph := buf.String()
```

**Benefits:**

- More efficient string building (O(n) instead of O(n²))
- Cleaner, more readable code
- Better performance for large network maps

**Acceptance Criteria:**

- [ ] Refactored to use strings.Builder
- [ ] Code is cleaner and more readable
- [ ] Benchmark shows performance improvement
- [ ] All existing tests pass

**Labels:** `improvement`, `performance`, `low-priority`

---

### Issue 8: Refactor CLI to Use Separate Subcommands per Platform

**Title:** Refactor CLI to use separate subcommands for each cloud platform

**Description:**

The current CLI uses a single command with a `--platform` flag. For better UX and clarity, refactor to use separate subcommands for each platform.

**Current Usage:**

```bash
ip2cr -platform=aws -ipaddr=1.2.3.4
ip2cr -platform=azure -ipaddr=1.2.3.4 -tenant-id=xyz
ip2cr -platform=gcp -ipaddr=1.2.3.4 -tenant-id=xyz
```

**Desired Usage:**

```bash
ip2cr aws -ipaddr=1.2.3.4
ip2cr azure -ipaddr=1.2.3.4 -tenant-id=xyz
ip2cr gcp -ipaddr=1.2.3.4 -tenant-id=xyz
```

**Related TODOs:**

- `cmd/root.go:103` - "TODO: change to separate subcommands per platform"
- `cmd/root.go:106` - "TODO: change to separate subcommands per service"

**Benefits:**

- Cleaner UX
- Platform-specific options are more obvious
- Better help documentation per platform

**Acceptance Criteria:**

- [ ] Separate subcommands for aws, azure, gcp
- [ ] Platform-specific flags only apply to relevant platforms
- [ ] Help documentation is clear and complete
- [ ] Backward compatibility or migration guide provided
- [ ] All existing tests pass and new tests added

**Labels:** `improvement`, `ux`, `refactoring`, `low-priority`

---

### Issue 9: Use errgroup.Group for Cleaner Goroutine Management

**Title:** Refactor concurrent search to use errgroup.Group

**Description:**

The current goroutine management in `search/search.go` uses manual `sync.WaitGroup` and channels. Using `golang.org/x/sync/errgroup` would provide cleaner error handling and goroutine management.

**Current Pattern:**

```go
matchingResourceBuffer := make(chan generalResource.Resource, 1)
var wg sync.WaitGroup
for _, acctID := range acctsToSearch {
    wg.Add(1)
    go search.runSearchWorker(matchingResourceBuffer, acctID, ...)
}
go func() {
    wg.Wait()
    close(matchingResourceBuffer)
}()
```

**Desired Pattern (using errgroup):**

```go
eg, ctx := errgroup.WithContext(context.Background())
for _, acctID := range acctsToSearch {
    acctID := acctID
    eg.Go(func() error {
        // search logic
        return nil
    })
}
if err := eg.Wait(); err != nil {
    // handle error
}
```

**Benefits:**

- Cleaner error propagation
- Built-in context support
- Reduces boilerplate code
- Prevents goroutine leaks

**Related Issue:** Should be done after Issue #4 (improve error handling)

**Acceptance Criteria:**

- [ ] Refactored to use errgroup.Group
- [ ] Error handling improved
- [ ] Code is cleaner and more maintainable
- [ ] Unit tests verify error collection
- [ ] All existing tests pass

**Labels:** `improvement`, `refactoring`, `low-priority`

---

### Issue 10: Improve Error Message Consistency

**Title:** Standardize error message formatting across codebase

**Description:**

Error messages are constructed inconsistently throughout the codebase using various patterns like `fmt.Sprintf()`, string concatenation, and `errors.New()`. Standardize on `fmt.Errorf()` for consistency and better debugging.

**Examples of Current Patterns:**

```go
// Pattern 1: fmt.Sprintf + errors.New
errorMsg := fmt.Sprintf("%s is not a supported platform for searching", search.Platform)
return matchingResource, errors.New(errorMsg)

// Pattern 2: String concatenation in log.Fatal
log.Fatal("error when connecting to ", search.Platform, ": ", err)

// Pattern 3: Multiple string joins
log.Info("starting resource search in AWS account: ", acctID, " ", acctAliases)
```

**Desired Pattern:**

```go
// Consistent use of fmt.Errorf
return matchingResource, fmt.Errorf("platform %q is not supported for searching", search.Platform)

// Structured logging with fields
log.WithFields(log.Fields{
    "platform": search.Platform,
    "error": err,
}).Error("failed to connect to platform")

// Proper formatting with context
log.WithFields(log.Fields{
    "account_id": acctID,
    "aliases": acctAliases,
}).Info("starting resource search in AWS account")
```

**Benefits:**

- Consistent formatting makes debugging easier
- Better structured logging
- Improved test readability
- Follows Go conventions

**Acceptance Criteria:**

- [ ] All error messages use `fmt.Errorf()` or proper error types
- [ ] Logging statements use structured fields where appropriate
- [ ] Error messages follow a consistent format
- [ ] All existing tests pass
- [ ] New tests added for error message formatting

**Labels:** `improvement`, `code-quality`, `low-priority`

---

## How to Use These Templates

1. Go to <https://github.com/magneticstain/ip-2-cloudresource/issues>
2. Click "New Issue"
3. Copy the **Title** into the title field
4. Copy the **Description** section into the body field
5. Add the **Labels** listed at the bottom
6. Submit the issue

You can create them in order (Priority 2 issues first) or all at once. Each issue is independent unless otherwise noted.
