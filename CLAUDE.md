# gofish (CoreWeave fork)

This is CoreWeave's fork of [stmcginnis/gofish](https://github.com/stmcginnis/gofish), a Go client library for DMTF Redfish and SNIA Swordfish. The module path is `github.com/coreweave/gofish`. Other CoreWeave services import it, so a change to an exported name, a field type or behavior is a compatibility change for them. The repo is public. Code, comments, commits and PRs must not name internal repos, services, tickets, incidents, hardware fleets or customers, and must not say which known weaknesses CoreWeave's copy still has. Describe a bug by the upstream issue or the vendor behavior.

## Upstream

- On 2026-04-15 the fork's tree was reset to upstream 8acdb3d (upstream #496). That is the last upstream commit before #497, which added `ComputerSystem.IPMIHostInterface`. Some BMCs send that field as a bare `true`, and the typed field then fails the whole ComputerSystem decode.
- Upstream then replaced the hand-written `common/`, `redfish/` and `swordfish/` packages with one generated `schemas/` package (#493, first released in v0.21.0). The fork keeps the hand-written layout. Upstream's AGENTS.md and `.claude/skills` describe the generator, `_custom.go` files and `make generate`, and none of that exists here. Their test rules for vendor workarounds still apply (see Tests).
- The two repos share no history after the reset. git's merge-base is upstream #434, so a merge from upstream would delete `redfish/` and `common/`. Never merge or cherry-pick from upstream. Port the fix by hand into the fork's file and cite the upstream PR, as #45, #47 and #55 did.
- Many upstream fixes after the reset repair regressions from the move to generated code (#498 to #509, #520, #521 and others, often described as restoring v0.20 behavior). The fork never had those bugs. Check the fork's code before porting one.
- `gh pr create` in a fork defaults to the parent repo. Pass `--repo coreweave/gofish`. Fork branches have been opened against upstream by mistake three times (upstream #525, #557, #558).
- Local clones fetch both remotes' tags. Fork tags are `v0.2.N`; upstream tags are `v0.2N.0`, so v0.27.0 is upstream's. List the fork's with `git ls-remote --tags origin`.
- Upstream's maintainer has declined two fork-only designs: per-call `WithContext` twins (upstream offers `ConnectContext` and `APIClient.WithContext`) and exported link fields or raw URIs. Don't propose them upstream.

## Layout

- `client.go`: `APIClient`, `ClientConfig`, connect, sessions, and the single request path that turns any status other than 200, 201, 202 or 204 into `*common.Error`. `MaxConcurrentRequests` defaults to 1 because some BMCs return 403 under parallel requests. Keep that default.
- `serviceroot.go`: `Service`, with every root link exported as `XLink common.Link`. Getters return `nil, nil` when a link is absent, so callers nil-check.
- `common/`: the `Entity` and `Resource` base types, the `Client` interface and `TestClient`, `Link` and `Links`, the generic fetchers (`GetObject`, `GetObjects`, `GetCollectionObjects`), `$expand` options in `odata.go`, the `Update` machinery, and `Error` and `CollectionError`.
- `redfish/` and `swordfish/`: one hand-written file per schema resource, with its test beside it.
- `oem/<vendor>/` (ami, dell, hpe, smc, zt): each type embeds a standard type and is built with `FromX(*redfish.X)`, which re-parses `RawData`. Core packages never import `oem/`.
- `examples/` is Markdown only. It and the README still import `github.com/stmcginnis/gofish`; new code uses the fork's path.

## Commands

```bash
go build ./... && go vet ./...
make test                                  # go test -v -cover -race over every package, as CI runs it
make lint                                  # golangci-lint v2.6.2 in Docker, config in .golangci.yaml
make fmt                                   # golangci-lint fmt in Docker
go test ./redfish/ -run TestChassis -count=1
```

- Skip `make generate`. It points at `tools/generator`, which the fork deleted.
- CI (`.github/workflows/check.yml`) builds and tests with Go 1.21 and runs golangci-lint. A newer local toolchain hides 1.21 problems: a stdlib API added after 1.21 compiles locally and fails in CI, and Go 1.21 ignores the `omitzero` tag option. Use neither.
- Lint limits that catch new code: funlen (100 lines, 50 statements), gocyclo 30, dupl 200, goconst (tests too), `nolintlint` (a `//nolint` that no longer suppresses anything fails), and `goheader`, which requires the `SPDX-License-Identifier: BSD-3-Clause` header on every Go file.
- The library has no dependencies and `go.sum` stays empty, tests included. depguard does not enforce this, so reviewers do.

## Writing resource code

- Every exported function or method that does I/O comes in two forms. `XWithContext(ctx, ...)` holds the body and calls only WithContext variants. `X(...)` is a one-line shim that passes `common.ContextOf(x.GetClient())`, or `common.ContextOf(c)` in a package function. Both get the same doc comment. Nothing checks the pairing, so add both. `Chassis.Update` and `GetChassis` in `redfish/chassis.go` show the shape.
- Getters and fetchers take `queryOpts ...common.QueryGroupOption` last and pass it on. `common.WithoutExpand()` turns `$expand` off for one call.
- A new method on `common.Client` goes on `APIClient` and `common.TestClient` in the same change, in both forms. It breaks consumers' own fakes; say so in the PR.
- Newer fork code exports links and actions: `BiosLink common.Link` with `json:"Bios"`, a `Links` struct, an `Actions` struct of `common.ActionTarget`, and `Settings common.Settings` with `json:"@Redfish.Settings"`. The accessor method stays and returns `nil, nil` when `IsZero()` is true. `redfish/computersystem.go` follows this. Older files such as `redfish/chassis.go` still copy links into unexported strings in a `type temp` UnmarshalJSON. Match the file you are in, and don't convert a whole file unless asked.
- Make an optional scalar a pointer (`*bool`, `*float64`) when absent and zero mean different things to a caller, as with `LocationIndicatorActive` and `Chassis.MaxPowerWatts`. That changes the field type for every consumer (see Compatibility).
- Go names use initialisms (`ID`, `OEM`, `URI`, `OData`) and the JSON tag carries the wire name (`json:"Id"`). Decoding matches keys without regard to case, but PATCH bodies use the tag, so a writable field whose Go name differs from the wire name needs a tag.
- Take types from the schema JSON: `integer` is an int type and `number` a float type. Some integer bounds don't fit in int64; Dell's AttributeRegistry `UpperBound` is a `big.Int` for that reason.
- Keep `RawData []byte` exported, and set it at the end of UnmarshalJSON, on any type that an `oem/` package or a caller re-parses.
- Library code returns errors. It doesn't print or log.

### Update()

`Update()` decodes `RawData` again, compares every exported field with the current struct and PATCHes the differences to `@odata.id`. If a changed field is missing from `readWriteFields`, the whole update fails with "X field is read only".

- `readWriteFields` lists JSON names (`RoleId`, not `RoleID`). Take them from the schema's read-write properties.
- A field the library fills after decode, such as a cache, needs `json:"-"` or the next `Update()` fails. `ComputerSystem.SupportedResetTypes` does this.

### Collections

- Collection fetches return the members that decoded plus a `*common.CollectionError` for the rest. Callers handle a partial slice with a non-nil error.
- `GetCollectionObjects` keeps document order (#45). `GetObjects`, which getters over `Links` arrays use, returns members in completion order.
- Member fetches run at most 3 at a time, and the client's `MaxConcurrentRequests` limit applies on top.
- `Members@odata.nextLink` pages are followed recursively, with `queryOpts` carried to each page.

## Vendor quirks and workarounds

A BMC that sends the wrong JSON type for one field (a number as a string, an object where the schema has an array, `[]` for an empty object) fails the decode of the whole resource, and with it every call that fetches it.

- For one vendor's firmware bug, keep the fork spec-shaped and fix the payload in the consuming service before it reaches gofish. #51 (an empty `EnergykWh` array from one AMI BMC) was closed for that reason. Add a workaround here when the spec allows the shape, when every consumer needs it, or when asked to.
- A workaround here decodes the spec shape first and falls back to the vendor shape only when that fails, so conformant services take the normal path. Use the `type temp` pattern so the decode doesn't recurse into itself. Copy an existing fallback: `parseMemberID` and `toFloat32` in `redfish/power.go`, the `MaxSpeedMHz` handling in `redfish/processor.go`, or `PartLocation.UnmarshalJSON` in `common/types.go`. Name the vendor, model and firmware in a comment on the workaround.
- `Entity.Headers()` sends the resource's ETag as `If-Match` on every POST and PATCH once one is known. Some firmware rejects it (Dell iDRAC10 returns 412 on Reset and InsertMedia) and some requires it (Lenovo, Huawei). Callers can switch it per object with `DisableEtagMatch` and `StripEtagQuotes`.
- With credentials, gofish creates a session, and the caller must `Logout` or the BMC hits its session cap. Don't add cookie auth: the spec forbids it, and it broke Inspur BMCs upstream.
- A new field copied from upstream's schemas can break every read of its resource. #53 added `IPMIHostInterface` and #54 removed it the same day, so v0.2.9 has it and nothing should pin v0.2.9. Before adding fields to ComputerSystem, Chassis, Manager or another resource that consumers read on every poll, check how BMCs send them. For `IPMIHostInterface`, port upstream's bool fallback (#575) first.

### Upstream fixes not ported yet

Search upstream's PRs before writing a fix. These landed upstream after the reset and are missing here. Delete a line when you port its fix.

- A custom transport loses HTTP/2 (#518).
- iDRAC10 returns 412 on action POSTs that carry If-Match (#510). Upstream #512, still open, sends If-Match only when the target URI is the entity's own. SetBoot on iDRAC10 needs the `@Redfish.Settings` target (#511).
- `UpdateBiosAttributes` panics on slice-valued attributes (#571).
- A `nextLink` that points back to an earlier page recurses without end (#572).
- `IPMIHostInterface` sent as a bool (#575).
- Numbers sent as strings: NVIDIA HGX `SensorArrayExcerpt.Reading` (#517), ZTE `Processor.Socket` (#533), Inspur `MemoryLocation.Channel` (#562), PCIe lane counts (#566). Lenovo XCC VirtualMedia without Insert or Eject actions (#530). Supermicro `SyslogServer` as an array (#553, #563).

## Tests

- Tests sit in the package they test (`package redfish`), so they can read unexported fields.
- Payloads go inline as raw-string vars in the `_test.go` file. No testdata directories, no `.json` files, no network; httptest servers are fine.
- A workaround's test decodes the vendor payload, trimmed to the fields involved, with a comment naming vendor, model and firmware. It asserts the corrected values and also decodes a spec-conformant payload for the same fields.
- A pointer change asserts the value when present and nil when absent (`TestMinimumChassis` in `redfish/chassis_test.go`).
- `common.TestClient` hands out `CustomReturnForActions` responses per HTTP verb in call order, and a call past the end panics. Collection members are fetched concurrently, so give members identical bodies or check captured calls as a set.
- `CapturedCalls()` stores the payload as `fmt.Sprintf("%v")` output, which only `strings.Contains` can match. To assert an exact PATCH body, wrap TestClient as `computerSystemBootClient` does in `redfish/computersystem_test.go`; `TestComputerSystemSettingsTargetPreservesDirectBootWrite` checks the verb, URL and body.
- A concurrency fix gets a deterministic test that fails without the fix under `-race`, as #45 has.
- Show that a new test fails without the change.

## Compatibility and releases

- Consumers pin release tags, so a change reaches them only after a tag and a bump in each consumer.
- A PR that changes an exported name, field type, signature or behavior says what breaks in its description and checks each consumer's use. The fork absorbs breaks with a follow-up PR in each consumer, not with deprecation shims.
- When a widely called function needs a new parameter, add a sibling (`XParams` with a struct, or `XWithTask`) and keep the old function as a wrapper, as upstream does.
- PR titles and commit subjects use Conventional Commits (`fix:`, `feat:`, `test:`, `ci:`). There is no `!` marker; breaks go in the description.
- main requires one approval, a fresh approval after the last push, and no force pushes. Merge, squash and rebase are all allowed.
- Releases are cut by hand as GitHub Releases with generated notes, tagged `v0.2.N`, usually when a consumer needs the change. Type changes ship as patch bumps. Don't tag or release unless asked.

## Review focus

These areas need a person's review even when the diff looks routine:

- `client.go`: transport, TLS, sessions, auth and concurrency.
- `common/entity.go`: `Update`, the PATCH diff, and `Headers()` with If-Match.
- `common/collection.go` and `GetObjects`: fan-out, ordering, pagination and partial errors.
- Exported types and fields on ComputerSystem, Chassis, Manager, Power, Thermal and their subsystems. Every consumer reads these in bulk, and one wrong type fails the whole resource.
- Missing `WithContext` twins or `queryOpts`, and `readWriteFields` lists.
