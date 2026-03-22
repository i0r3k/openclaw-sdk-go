# Changelog

All notable changes to this project will be documented in this file.

## Unreleased



### 🚀 Features

- Add common types and error type hierarchy ([1396bbb](1396bbbde374deea6939e8ddcec9eecb1c4d3387)) - (Lin Yang)
- Add Logger interface with context support ([d1d28ba](d1d28baf06bd832fe64f791e3a97518ce0076e5e)) - (Lin Yang)
- Add auth module with CredentialsProvider and AuthHandler ([16a2b15](16a2b15c1d6310994adfbf94ea336aebf017947b)) - (Lin Yang)
- Add protocol module with types and validation ([6040de1](6040de1e2888ac0b750e3c627b040dece09807f4)) - (Lin Yang)
- Add WebSocket transport module ([9d04235](9d042352b058fad02d23206c66d013cdacb7ca71)) - (Lin Yang)
- Add connection module with state machine and policies ([f066db5](f066db53751ff14706776be68f66724d791b078d)) - (Lin Yang)
- Add events module with tick monitor and gap detector ([044d7be](044d7bea339428cbc84fd831cda6e70a1f896eee)) - (Lin Yang)
- Add managers module with event, request, connection, and reconnect managers ([672bf8b](672bf8b3358380eb4086bd24b0dc653a990315ba)) - (Lin Yang)
- Add timeout manager ([1e1161b](1e1161b4a971b2823dd4249677bf769e5cd51ca6)) - (Lin Yang)
- Add main client with options and reorganize package structure ([162e528](162e528f92822375e8b87e1bd04b551e700c2b55)) - (Lin Yang)
- Add CLI example ([c95b9dc](c95b9dc9789612172bb28a0631547f4f18101c4b)) - (Lin Yang)
- Add WebSocket echo server example ([758ae6f](758ae6fb02b5d13802eaba0332a46386ae3454fc)) - (Lin Yang)
- Enhance TLS certificate validation with comprehensive security checks ([5fa31b4](5fa31b4023f6ef37d1a23798690e0d90b9c49da4)) - (Lin Yang)
- Implement new features of TypeScript to Go SDK migration ([3d24935](3d24935cc4aac393acf48b8d3240c2231e09ff7c)) - (Lin Yang)
- Add context support to Dial operation ([b53e978](b53e9783defff4a9bdb9868fecd2b5fe4c55bf53)) - (Lin Yang)
- Add backpressure timeout for EventManager Emit operations ([3148bfc](3148bfc6e334caeecfa215e185450acca5cea0fb)) - (Lin Yang)
- Add payload size validation against server policy ([9c93c48](9c93c480fab36a4f38ae0fa115b97cd670080cfc)) - (Lin Yang)
- Integrate git-cliff for structured release notes generation ([1bff317](1bff317d110368f8608154dd09a3d84c29fc6749)) - (Lin Yang)

### 🐛 Bug Fixes

- Prevent memory leak in ReconnectManager by using time.NewTimer ([da1d24a](da1d24ad4497eb35b8460b3d051f283f53cb80d4)) - (Lin Yang)
- Remove infinite loop in ProtocolNegotiator and simplify version matching ([47f3a73](47f3a73e3f034617e27a5752c1655d76f37c466e)) - (Lin Yang)
- Properly handle errors in connection manager and transport ([d204694](d2046944a29c892995ee345b84384c1c35619fd6)) - (Lin Yang)
- Properly load TLS certificates in transport layer ([373d972](373d97266e0173df6269f451607671bafcb5cb4a)) - (Lin Yang)
- Resolve Timer reset race condition in TickMonitor ([d1612f4](d1612f40292937ede2e40e3391e2dd47020d3e86)) - (Lin Yang)
- Log panics in EventManager handlers instead of silent discard ([d070973](d0709736d2c6eb6b3135b51ee37e9eed1d6562b7)) - (Lin Yang)
- Replace unsafe.Pointer with atomic counter in EventManager ([ef79823](ef7982355bc560049f178151c49e237053255a4c)) - (Lin Yang)
- Validate credential values in auth package ([c22a551](c22a5517814b80a8d74c42876261ae3305ecd90b)) - (Lin Yang)
- Enhance method name validation with regex and comprehensive tests ([29f9eb1](29f9eb1c5a85e1244f6c6e860acdd52c19aeacd1)) - (Lin Yang)
- Replace type assertions with json.Unmarshal to prevent runtime panics ([b5e39db](b5e39dba63a5c38958d6fd8b64cdce7d8c044f3e)) - (Lin Yang)
- Implement performHandshake and fix state transition path ([aaacfcc](aaacfccc345acfeabec9556377eeb87eb41ae5a1)) - (Lin Yang)
- Reconnect uses stored params instead of raw Connect ([5f739dc](5f739dc16be8f542e482906731d7ee254c278088)) - (Lin Yang)
- Use strconv.Itoa for protocol version to prevent overflow ([9524606](9524606ffcd4fdaf16d4f1b80fc6c4fb4642f11e)) - (Lin Yang)
- Add closed flag to prevent RequestManager double-close ([6a9023b](6a9023b0d903d1c5b2bc5c1984e05fee1e173833)) - (Lin Yang)
- Use write lock in GetStaleDuration to prevent data race ([0bae076](0bae076286423d151cef725f2b7ef1c07b8bb9a2)) - (Lin Yang)
- Add background goroutine for automatic stale detection ([8401c39](8401c39df4a6e001689f6bf269ef9d1eedee9e15)) - (Lin Yang)
- Simplify IsRequestError and fix Unwrap chain ([a7c471a](a7c471a50431be9332d526629447988c7120fa78)) - (Lin Yang)
- Use crypto/rand for request ID generation ([d30949c](d30949c8754a2f75dc5bf80eccb3028d38b2204d)) - (Lin Yang)
- Add WithClientID option and fix tests missing ClientID ([8a78b06](8a78b061175c91c190d98031dae46f5f8427f837)) - (Lin Yang)
- Make channel buffer size configurable ([e42c623](e42c6237d19187f4d6e2a8d560e26cc4ac5d54cc)) - (Lin Yang)
- Log warning when events are dropped due to full channel ([baff296](baff296143c75e3b68f57eb6259ad737ea9f7fa3)) - (Lin Yang)
- Resolve race conditions in request and event managers ([5853e8e](5853e8ef81097be40aa035a9197bc08a3a2adc36)) - (Lin Yang)
- Log protocol negotiation errors instead of silently discarding ([0652469](0652469b16cba3d58070967c2a7b841cf9ec48c5)) - (Lin Yang)
- Add security warning when InsecureSkipVerify is enabled ([e438a5f](e438a5ffb93e83d72952dcaf54cfe8bdefce9f94)) - (Lin Yang)
- Migrate golangci-lint config to v2 format ([d980789](d98078918532f5fcde5728fbef73509ffe366f4d)) - (Lin Yang)
- Remove invalid --all flag from git-cliff command ([be83719](be83719db83a15bd904dfc4e1eb262f3f8d05c31)) - (Lin Yang)
- Skip build step for library project ([022a142](022a1426423af3eb0c65789c25deaf69a021a1ef)) - (Lin Yang)
- Rename build to builds in goreleaser config ([2194e31](2194e31bd198e3ca30b6c9a60097929e9c253f3e)) - (Lin Yang)
- Fixup! fix(ci): rename build to builds in goreleaser config ([fd15547](fd155473e890e0a57c9fb70e898686fab4ab72f0)) - (Lin Yang)


### ⚡ Performance

- Eliminate redundant JSON marshal and slice allocation ([81b701f](81b701f447ad7d92ef155db49d4a355d303e1b26)) - (Lin Yang)
- Replace time.After with time.NewTimer+Reset in EventManager.Emit ([bac682c](bac682c2db2fbd2b90eaa338f0a9c94c6476d377)) - (Lin Yang)
- Eliminate heap allocation in generateRequestID ([26b8d90](26b8d90a4cbab91a531dc6650114de327b674884)) - (Lin Yang)

### ♻️ Refactor

- Move source files to pkg/openclaw directory ([be15cff](be15cffb1fd981918c70694f0d4c4bee1a8b04fe)) - (Lin Yang)
- Move utils package from pkg/openclaw/utils to pkg/utils ([b5e5bee](b5e5bee3a4a5de5f56260449f28f1b00d6861c1a)) - (Lin Yang)
- Move openclaw package from pkg/openclaw to pkg ([b748cb7](b748cb737656c9e2d82c5e845376efd1f681a0ac)) - (Lin Yang)
- Consolidate re-exports in client.go, move types to pkg/types/ ([e5fffa6](e5fffa6cbf368662d0d3424ea2629685bed782a2)) - (Lin Yang)
- Clean up unused code and fix misleading comment ([e91520a](e91520a5547015b6155b31c14003e833f0447b70)) - (Lin Yang)
- Remove unused stateMachine field ([e97a82c](e97a82ce5a4a9cb9e0a817a0424086086c2238b8)) - (Lin Yang)
- Split api_params.go into domain-specific files ([26cfc60](26cfc60f4d63211d065c89ae560bc86597cf5921)) - (Lin Yang)
- Remove dead setupConnectionHandlers placeholder ([6093aae](6093aaeac930d0b2c0feb071e23856270147a830)) - (Lin Yang)

### ✅ Testing

- Add comprehensive error type tests ([c04a9a6](c04a9a65fc5e0cd832c8f51727d6fb67bd53864a)) - (Lin Yang)
- Enhance test coverage and fix bugs ([4ebfe4a](4ebfe4ab96b182500a3ad37b83fe20ecda479ffa)) - (Lin Yang)
- Add comprehensive tests for all 8 API namespaces ([f22698e](f22698e1466839d8dbf4f896e208d86ab6da623b)) - (Lin Yang)
- Fix MalformedJSON no-op test to actually parse JSON ([24c930c](24c930c9ec90fe79a755198045a1a91c4a3e511a)) - (Lin Yang)
- Improve coverage for types, api, managers, utils, and client packages ([bd53c2d](bd53c2da8ecff05ccfd3bb07ea7bc5c88ef1c56d)) - (Lin Yang)



### 📖 Documentation

- Add TypeScript to Go migration design spec ([2c4de56](2c4de56229b58f4955e36f23e1dd89a1c7371973)) - (Lin Yang)
- Add implementation plan with 10 phases ([68e5644](68e56443f516632e3599723e8b9116862c7d82c3)) - (Lin Yang)
- Improve Phase 1-2 implementation plans ([e864988](e86498889cb022da0bc9a2301e7bb61c0b3752ed)) - (Lin Yang)
- Improve Phase 3 protocol module plan ([add24e1](add24e142a5d9b6082abf0f4d742a49df726bfd6)) - (Lin Yang)
- Improve Phase 4 transport module plan ([3326f1d](3326f1de79febeb3c22ab7fa029545e3bbc81d36)) - (Lin Yang)
- Improve Phase 5 connection module plan ([16f0ff9](16f0ff959dbdcc71d401de0b230e3b0f53e2f865)) - (Lin Yang)
- Improve Phase 6 events module plan ([808a74b](808a74b4190935e0d1aa0a7a17b6f6bc05e050d3)) - (Lin Yang)
- Improve Phase 7 managers module plan ([18fc763](18fc763c141ca07ddc71c6e5f5d31ce29bec8d6a)) - (Lin Yang)
- Improve Phase 7 managers module plan ([9251be1](9251be18d3fde50090790b4091123e4696a47648)) - (Lin Yang)
- Add timeout validation to Phase 8 utils module plan ([f1d2abe](f1d2abebb17c045f321ea6133f312522f8216c05)) - (Lin Yang)
- Implement Phase 9 main client with thread-safe operations ([6deda27](6deda27b1fd3fd85b021426f4719c332e742058e)) - (Lin Yang)
- Fix import paths and cross-phase API compatibility ([cddfd4d](cddfd4d7231dab75257048f4d9da809e77588107)) - (Lin Yang)
- Update all phase plans to use pkg/openclaw directory structure ([8dcac5e](8dcac5e9c57433d5116db8d4b1f00ffc5341c978)) - (Lin Yang)
- Add CLAUDE.md with project overview and development commands ([9f3d6d5](9f3d6d5c0a89cb7be32765b8b3bdd7ec65bb6efc)) - (Lin Yang)
- Add comprehensive README with installation, usage, and API documentation ([f205051](f205051a193bcb5bc9df0bccbebeccc4764c939a)) - (Lin Yang)
- Update README with badges and project description ([4a61d69](4a61d69a55b38a5637fe4b5a4919551417b055e2)) - (Lin Yang)
- Add comprehensive documentation comments to all non-test Go files ([253f510](253f5104c8523e23a4e79dc8eac9066586c63e07)) - (Lin Yang)
- Document dual ErrorShape design intent ([8c8543a](8c8543a4d5b131e072bd0ae63df4bc6f5b353af4)) - (Lin Yang)
- Improve CLAUDE.md with Key Files and Gotchas sections ([6d858d8](6d858d874ff220991a9309a7548e6085542b7a6f)) - (Lin Yang)
- Clarify CRL/OCSP stub with actionable TODO ([66620a7](66620a703108ee829ced44c4086b3b6de5297d07)) - (Lin Yang)
- Add Codecov coverage badge ([803e7ef](803e7ef60d56799c0e52b70b55299355b7247a0e)) - (Lin Yang)
- Update with new features and API additions since v1.0 ([8fb6042](8fb604231bb29f7bc87e19716a0b8d82c2dbefa7)) - (Lin Yang)

### 🔧 Miscellaneous Tasks

- Add GitHub Actions workflows ([88f31fe](88f31fef50fca21765641540e86739d7fe6cb345)) - (Lin Yang)
- Modify CI workflow for Go version and fail-fast ([1abdb9e](1abdb9ea1518365853e58e64e3caeba2e5cf4bee)) - (Lin Yang)



<!-- generated by git-cliff -->
