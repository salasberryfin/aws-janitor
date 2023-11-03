# AWS Janitor

A GitHub Action to cleanup AWS resources that have exceeded a TTL.

> By default the action will not perform the delete (i.e. it will be a dry-run). You need to explicitly set commit to `true`.

It supports cleaning up the following services:

- EKS Clusters
- Auto Scaling Groups

## Inputs

| Name              | Required | Description                                                                            |
| ----------------- | -------- | -------------------------------------------------------------------------------------- |
| regions           | Y        | A comma seperated list of regions to clean resources in. You can use * for all regions |
| allow-all-regions | N        | Set to true if use * from regions.                                                     |
| ttl               | Y        | The duration that a resource can live for. For example, use 24h for 1 day.             |
| commit            | N        | Whether to perform the delete. Defaults to `false` which is a dry run                  |

## Example Usage

```yaml
jobs:
  cleanup:
    runs-on: ubuntu-latest
    name: Cleanup resource groups
    steps:
      - name: Cleanup
        uses: rancher-sandbox/aws-janitor@v0.1.0
        with:
            regions: eu-west-1
            ttl: 168h
        env:
            AWS_ACCESS_KEY_ID: {{secrets.AWS_ACCESS_KEY_ID}}
            AWS_SECRET_ACCESS_KEY: {{secrets.AWS_SECRET_ACCESS_KEY}}
```

## Implementation Notes

It currently assumes that an instance of a service will have some form of creation date. This means that the implementation can be simpler as it doesn't need to adopt a "mark & sweep" pattern that requires saving state between runs of the action.
