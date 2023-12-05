#!/usr/bin/env bash
set -xe

function generate_upgrader_crds() {
  local sources_path="crds/sources/rancher/upgrader/plan"
  crd2pulumi --goPath crds/generated/rancher ${sources_path}/*.yaml --force
}

generate_upgrader_crds
