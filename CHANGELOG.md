# Changelog

## [0.2.0](https://github.com/woodleighschool/freshservice-label/compare/0.1.3...0.2.0) (2026-08-22)


### ⚠ BREAKING CHANGES

* webhook payload fields have been replaced by the generic label schema, and the embedded Woodleigh template is no longer used.

### Features

* **container:** update image golang (1.26.6 → 1.27.0) ([#18](https://github.com/woodleighschool/freshservice-label/issues/18)) ([0ba0ce4](https://github.com/woodleighschool/freshservice-label/commit/0ba0ce4164552d64acab3aa359a061de5ee2db6a))
* **deps:** update module golang.org/x/image (v0.44.0 → v0.45.0) ([#16](https://github.com/woodleighschool/freshservice-label/issues/16)) ([cf8df03](https://github.com/woodleighschool/freshservice-label/commit/cf8df0326d01adc871a74c06e84a23357529bcf9))
* move label content out of the renderer ([bfd22d4](https://github.com/woodleighschool/freshservice-label/commit/bfd22d41db0e9021ce00f30cc04e6adc6212d795))


### Bug Fixes

* **go:** update module github.com/go-chi/chi/v5 (v5.3.1 → v5.3.2) ([#22](https://github.com/woodleighschool/freshservice-label/issues/22)) ([207614c](https://github.com/woodleighschool/freshservice-label/commit/207614c2c6467de488600f9558b315368cc0f7c8))
* **renovate:** wait for complete toolchain groups ([d5586b2](https://github.com/woodleighschool/freshservice-label/commit/d5586b267b5c805c2c57941e6415fcc61d695f7e))
* **tooling:** group toolchain updates ([563a2eb](https://github.com/woodleighschool/freshservice-label/commit/563a2eb9b81234d3798d977a86696efa3bab4ca6))


### Code Refactoring

* align configuration and logging ([f543469](https://github.com/woodleighschool/freshservice-label/commit/f54346932f1394ab36f77e18486c6534a14e775a))


### Build System

* write binary to repository root ([2d57a65](https://github.com/woodleighschool/freshservice-label/commit/2d57a65f84e1377e4e89a67ac0ebfeeb1f2aa31e))


### Continuous Integration

* **github-action:** Update action docker/github-builder (v1.15.0 → v1.16.0) ([2214ca4](https://github.com/woodleighschool/freshservice-label/commit/2214ca4375c96bc6c5a1f27011e0d71fd8423cde))
* **github-action:** Update action home-operations/.github/actions/workflow-lint (v1.0.2 → v1.0.3) ([#19](https://github.com/woodleighschool/freshservice-label/issues/19)) ([06cb329](https://github.com/woodleighschool/freshservice-label/commit/06cb32964a0edc7c27e9fedcfa79bfab66d9987a))
* **github-action:** Update action jdx/mise-action (v4.2.3 → v4.2.4) ([a76be44](https://github.com/woodleighschool/freshservice-label/commit/a76be4449a9067dc774002fe43c06ab0db7366a7))
* **github-action:** Update action jdx/mise-action (v4.2.4 → v4.2.5) ([343342d](https://github.com/woodleighschool/freshservice-label/commit/343342dd3d52f6ad9fd850ab036d2d668793bb90))
* standardize workflows and add govulncheck ([42c58b8](https://github.com/woodleighschool/freshservice-label/commit/42c58b8470ecb046846f58a402e7e52dfbf2f1a0))
* sync shared repository tooling ([49cfa8e](https://github.com/woodleighschool/freshservice-label/commit/49cfa8e8bc49f8f0715dd164ea433dd5707673ea))


### Miscellaneous Chores

* align ignore rules ([f4047ee](https://github.com/woodleighschool/freshservice-label/commit/f4047ee0f9e76bcccea6f7987d8bcd838b84cd87))
* align repository conventions ([32d3125](https://github.com/woodleighschool/freshservice-label/commit/32d31252bcfbf5de20d7c64abdc55fd3eda72925))
* align repository conventions ([d0cc680](https://github.com/woodleighschool/freshservice-label/commit/d0cc68074e882707b19ba74bcc831120bf4fd415))
* **go:** update toolchain to 1.27 ([109d30b](https://github.com/woodleighschool/freshservice-label/commit/109d30b199a98e1aa14cee589c09878ab09c707d))
* **mise:** update tool golangci-lint (2.13.0 → 2.13.1) ([#20](https://github.com/woodleighschool/freshservice-label/issues/20)) ([75d9c1d](https://github.com/woodleighschool/freshservice-label/commit/75d9c1de9b0abde133f8945f4eb94ee11e0a689c))
* **mise:** update tool lefthook (2.1.10 → 2.1.11) ([#21](https://github.com/woodleighschool/freshservice-label/issues/21)) ([a2400a1](https://github.com/woodleighschool/freshservice-label/commit/a2400a15a81d5d1e45cd1478091e99e30125aeeb))
* **mise:** Update tool oxfmt (0.62.0 → 0.63.0) ([#15](https://github.com/woodleighschool/freshservice-label/issues/15)) ([47f8bfc](https://github.com/woodleighschool/freshservice-label/commit/47f8bfc25a77c61fbf36d8c346c1a72203e10809))
* **release-please:** sync configuration ([cbac61a](https://github.com/woodleighschool/freshservice-label/commit/cbac61ace7e0d4811c7bb28e4f0b8e9f45aa4464))
* **tooling:** sync shared configuration ([57bca04](https://github.com/woodleighschool/freshservice-label/commit/57bca04123176d57f251ec3b6e1b4d7755388e04))

## [0.1.3](https://github.com/woodleighschool/freshservice-label/compare/0.1.2...0.1.3) (2026-08-04)


### Bug Fixes

* **ci:** disable automatic mise installs ([88ee45e](https://github.com/woodleighschool/freshservice-label/commit/88ee45ed6f00999830776f06726c4bc23a88e551))
* resolve Go lint findings ([9ab827b](https://github.com/woodleighschool/freshservice-label/commit/9ab827b45fc9e2edc12945bd8e30c429c4c824af))

## [0.1.2](https://github.com/woodleighschool/freshservice-label/compare/v0.1.1...0.1.2) (2026-07-28)


### Features

* **deps:** update module github.com/go-chi/chi/v5 (v5.2.5 → v5.3.1) ([#3](https://github.com/woodleighschool/freshservice-label/issues/3)) ([6774f29](https://github.com/woodleighschool/freshservice-label/commit/6774f29dda5321c9cecb5117a8245a96e0e08828))
* **deps:** update module golang.org/x/image (v0.41.0 → v0.44.0) ([#4](https://github.com/woodleighschool/freshservice-label/issues/4)) ([c71d3cf](https://github.com/woodleighschool/freshservice-label/commit/c71d3cf5e6c5b24194c6914e52da6bebe6fe91b3))


### Bug Fixes

* **ci:** align repository tooling ([b32f7de](https://github.com/woodleighschool/freshservice-label/commit/b32f7dead37f82eaf0e89ba334482a9a3ae5efd7))
