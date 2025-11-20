# Droplet MCP Tools Implementation Plan

This document tracks the implementation progress for exposing all droplet-related functionality from the `godo` library via MCP tools.

## Status Summary

**Completed Steps:** 5/7  
**Current Step:** Step 6 - Action Retrieval by URI  
**Test Status:** ✅ All tests passing (51 test cases)

---

## Implementation Steps

### ✅ Step 1: Expose Existing Neighbor Tool (COMPLETED)

**Status:** ✅ Complete  
**Date Completed:** Current session

**Changes Made:**
- Registered `getDropletNeighbors` handler in `DropletTool.Tools()` as `droplet-neighbors`
- Added comprehensive test coverage in `droplet_tools_test.go`
- Updated `README.md` with tool documentation

**Tools Added:**
- `droplet-neighbors` - List droplets that share the same physical host

**Test Results:**
- ✅ `TestDropletTool_getDropletNeighbors` - 2 test cases passing

---

### ✅ Step 2: Expand Droplet Listing & Bulk Management (COMPLETED)

**Status:** ✅ Complete  
**Date Completed:** Current session

**Changes Made:**
- Added helper functions: `formatDropletSummaries`, `getNumberArg`, `parseStringArray`, `parseSSHKeys`, `parseTags`
- Implemented new handlers:
  - `listDropletsByName` - Filter by exact name
  - `listDropletsByTag` - Filter by tag
  - `listDropletsWithGPUs` - Filter GPU-enabled droplets
  - `createMultipleDroplets` - Bulk create with same config
  - `deleteDropletsByTag` - Bulk delete by tag
- Registered all new tools in `Tools()` method
- Updated `README.md` with full documentation
- Added comprehensive test coverage for all new handlers

**Tools Added:**
- `droplet-list-gpus` - List GPU-enabled droplets
- `droplet-list-by-name` - List droplets by exact name match
- `droplet-list-by-tag` - List droplets by tag
- `droplet-create-multiple` - Create multiple droplets at once
- `droplet-delete-by-tag` - Delete all droplets with a tag

**Test Results:**
- ✅ `TestDropletTool_listDropletsWithGPUs` - 2 test cases passing
- ✅ `TestDropletTool_listDropletsByName` - 3 test cases passing
- ✅ `TestDropletTool_listDropletsByTag` - 3 test cases passing
- ✅ `TestDropletTool_createMultipleDroplets` - 3 test cases passing
- ✅ `TestDropletTool_deleteDropletsByTag` - 3 test cases passing

**godo Methods Exposed:**
- ✅ `Droplets.ListWithGPUs`
- ✅ `Droplets.ListByName`
- ✅ `Droplets.ListByTag`
- ✅ `Droplets.CreateMultiple`
- ✅ `Droplets.DeleteByTag`

---

### ✅ Step 3: Expose Snapshot/Backup/Action Listings (COMPLETED)

**Status:** ✅ Complete  
**Date Completed:** Current session

**Changes Made:**
- Added handlers for `client.Droplets.Snapshots`, `Backups`, and `Actions` methods
- Implemented pagination support for all three listing endpoints
- Registered new tools: `droplet-snapshots`, `droplet-backups`, `droplet-actions-list`
- Updated `README.md` with full documentation
- Added comprehensive test coverage (9 new test cases)

**Tools Added:**
- `droplet-snapshots` - List snapshots for a droplet
- `droplet-backups` - List backups for a droplet
- `droplet-actions-list` - List all actions for a droplet

**godo Methods Exposed:**
- ✅ `Droplets.Snapshots(ctx, dropletID, *ListOptions)`
- ✅ `Droplets.Backups(ctx, dropletID, *ListOptions)`
- ✅ `Droplets.Actions(ctx, dropletID, *ListOptions)`

**Test Results:**
- ✅ `TestDropletTool_listDropletSnapshots` - 3 test cases passing
- ✅ `TestDropletTool_listDropletBackups` - 3 test cases passing
- ✅ `TestDropletTool_listDropletActions` - 3 test cases passing

---

### ✅ Step 4: Backup Policy Visibility & Control (COMPLETED)

**Status:** ✅ Complete  
**Date Completed:** Current session

**Changes Made:**
- Implemented handlers for backup policy inspection:
  - `getDropletBackupPolicy` -> exposes `Droplets.GetBackupPolicy`
  - `listBackupPolicies` -> exposes `Droplets.ListBackupPolicies`
  - `listSupportedBackupPolicies` -> exposes `Droplets.ListSupportedBackupPolicies`
- Implemented action handlers:
  - `enableBackupsWithPolicy` -> exposes `DropletActions.EnableBackupsWithPolicy`
  - `changeBackupPolicy` -> exposes `DropletActions.ChangeBackupPolicy`
- Registered all new tools in the respective `Tools()` methods
- Added comprehensive tests for each new handler
- Updated repository mocks to include new godo methods

**Tools Added:**
- `droplet-backup-policy-get` - Get the backup policy for a droplet
- `droplet-backup-policies-list` - List all backup policies
- `droplet-backup-policies-supported` - List supported policy options
- `droplet-enable-backups-with-policy` - Enable backups with a structured policy
- `droplet-change-backup-policy` - Change a droplet's backup policy

**godo Methods Exposed:**
- ✅ `Droplets.GetBackupPolicy`
- ✅ `Droplets.ListBackupPolicies`
- ✅ `Droplets.ListSupportedBackupPolicies`
- ✅ `DropletActions.EnableBackupsWithPolicy`
- ✅ `DropletActions.ChangeBackupPolicy`

**Test Results:**
- ✅ `TestDropletTool_getDropletBackupPolicy` - 3 test cases passing
- ✅ `TestDropletTool_listBackupPolicies` - 2 test cases passing
- ✅ `TestDropletTool_listSupportedBackupPolicies` - 2 test cases passing
- ✅ `TestDropletActionsTool_enableBackupsWithPolicy` - 3 test cases passing
- ✅ `TestDropletActionsTool_changeBackupPolicy` - 3 test cases passing

---

### ✅ Step 5: Enhance Droplet Create Arguments & Add More Request Fields (COMPLETED)

**Status:** ✅ Complete  
**Date Completed:** Current session

**Changes Made:**
- Extended the `droplet-create` handler to accept and propagate additional `godo.DropletCreateRequest` fields:
  - `IPv6` — enable IPv6 networking
  - `VPCUUID` — specify VPC UUID for placement
  - `UserData` — accept cloud-init user data
  - `Volumes` — accept an array of volume IDs (converted into appropriate godo types)
  - `WithDropletAgent` — optional boolean; now handled as a pointer (only set when the argument is provided)
  - `BackupPolicy` — accepts JSON-encoded `DropletBackupPolicyRequest` and unmarshals into request
- Extended the `createMultipleDroplets` handler to support the same additional fields (uses `godo.DropletMultiCreateRequest`)
- Added helper functions:
  - `parseStringArray` — validate and coerce interface arrays to []string
  - `parseSSHKeys` — handle numeric IDs and string fingerprints for SSH keys
  - `parseTags` — normalize tag arrays
  - `getNumberArg` — numeric argument helper with defaults
  - `formatDropletSummaries` — produce consistent JSON summaries for list/create-multiple responses
- Made `WithDropletAgent` handling safe: we create a pointer only when the argument exists so tests/mocks that don't set the field continue to match expected request shapes
- Implemented robust `BackupPolicy` parsing with clear error handling (returns tool error on invalid JSON)
- Updated and added tests to exercise the enhanced fields and ensure backward compatibility:
  - `TestDropletTool_createDroplet_with_extra_fields`
  - `TestDropletTool_createMultipleDroplets_with_extra_fields`
  - Existing create tests still pass without providing the new optional fields

**Tools Added / Updated:**
- Updated `droplet-create` to accept new optional arguments
- Updated `droplet-create-multiple` to accept new optional arguments
- No new tool names were required — existing tools were extended

**Test Results:**
- ✅ `TestDropletTool_createDroplet_with_extra_fields` — passing
- ✅ `TestDropletTool_createMultipleDroplets_with_extra_fields` — passing
- All existing droplet tests remain green; full droplet package test suite passes.

**godo Methods / Behaviors Exposed:**
- Creation flows now populate additional create-time fields on `godo.DropletCreateRequest` and `godo.DropletMultiCreateRequest` when provided, without changing the default behavior when omitted.

**Notes:**
- Mocks in tests use type-assignable matchers for the enhanced create requests to avoid brittle exact-match expectations when optional fields are omitted.
- The implementation favors being conservative about setting pointer fields on request structs — only provide pointers when the corresponding argument is present.

---

### ⏳ Step 6: Action Retrieval by URI (PENDING)

**Status:** ⏳ Pending  
**Priority:** Low

**Planned Changes:**
- Add handler that calls `DropletActions.GetByURI` to retrieve an action by its URI (useful for webhook-delivered action links).
- Register the tool in `DropletActionsTool.Tools()` as `droplet-action-by-uri`.
- Implement argument parsing to accept a single `URI` string argument and validate it.
- Add unit tests covering:
  - Successful retrieval by URI
  - API error path
  - Missing/invalid URI argument
- Update `README.md` and package mocks to include the new godo method if not already present.

**Tools to Add:**
- `droplet-action-by-uri` - Get a droplet action by its resource URI

**godo Methods to Expose:**
- `DropletActions.GetByURI`

---

### ⏳ Step 7: Additional Enhancements (PENDING)

**Status:** ⏳ Pending  
**Priority:** Low

**Planned Changes:**
- Add `ListAssociatedResourcesForDeletion` tool for safer cleanup workflows
- Consider adding `NeighborsByTag` if API supports it
- Review and optimize response formatting
- Add any missing edge case handling

**Tools to Consider:**
- `droplet-associated-resources` - List resources that would be deleted with droplet

**godo Methods to Expose:**
- `Droplets.ListAssociatedResourcesForDeletion`

---

## Test Coverage Summary

**Total Test Cases:** 51  
**Passing:** 51 ✅  
**Failing:** 0  
**Coverage Areas:**
- Droplet CRUD operations
- Droplet listing and filtering
- Bulk operations (create multiple, delete by tag)
- Neighbor discovery
- All droplet actions (power, reboot, resize, etc.)
- Tag-based bulk actions
- Image management
- Size listing
- Snapshot listings
- Backup listings
- Action listings

---

## Next Steps

1. **Immediate:** Complete Step 6 (Action Retrieval by URI)
2. **Short-term:** Implement and test enhanced droplet create fields (IPv6, VPCUUID, UserData, Volumes, WithDropletAgent, BackupPolicy)
3. **Medium-term:** Implement and test additional action retrieval and URI-based helpers (e.g., `GetByURI`)
4. **Long-term:** Complete Step 7 (Additional Enhancements)

---

## Notes

- All tools use argument-based input (no resource URIs)
- Pagination is supported where applicable via `Page` and `PerPage` arguments
- Error handling follows consistent patterns across all tools
- Response formatting uses JSON for easy parsing
- Test coverage follows existing patterns with success and error cases
