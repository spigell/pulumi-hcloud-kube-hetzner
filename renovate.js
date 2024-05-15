module.exports = {
  branchPrefix: 'test-renovate/',
  enabledManagers: ["custom.regex"],
  username: 'renovate-release',
  gitAuthor: 'Renovate Bot <bot@renovateapp.com>',
  onboarding: false,
  platform: 'github',
  includeForks: true,
  dryRun: 'full',
  customManagers: [
    {
      customType: "regex",
      fileMatch: ["examples/k3s-private-non-ha-current-latest-channel.yaml"],
      matchStrings: ["target-channel: (?<currentValue>\\S+)"],
      depNameTemplate: "k3s",
      versioningTemplate: "semver-coerced",
      datasourceTemplate: "custom.k3s"
    }
  ],
  customDatasources: {
    k3s: {
      defaultRegistryUrlTemplate: "https://update.k3s.io/v1-release/channels",
      transformTemplates: [
      	`{
          "releases": [
            {
            	"version": data[id = 'latest'].latest ~> $replace(/(v1\.[0-9][0-9])(.*)/, "$1")
            }
          ]
        }`
      ]
    }
  }
};