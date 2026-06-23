# Pulumi

> AWS serverless stack — API Gateway, Lambda, and static website.

## In this repo

| Path | Purpose |
|------|---------|
| `pulumi/aws/prod/main.go` | Stack entry (Go) |
| `pulumi/aws/prod/function/handler.py` | Lambda handler (Python) |
| `pulumi/aws/prod/www/index.html` | Static site content |
| `pulumi/aws/prod/Pulumi.yaml` | Project metadata |
| `pulumi/aws/prod/Pulumi.prod.yaml` | Stack config |

## Quick start

```bash
cd pulumi/aws/prod
pulumi login
pulumi stack select prod   # or create
pulumi up
```

Requires AWS credentials configured for the target account.

## Configuration

Stack provisions:

- IAM role for Lambda (basic execution)
- Python Lambda function (`handler.py`)
- API Gateway HTTP API route
- Static `www/` assets

### Stack config

Edit `Pulumi.prod.yaml` for region, tags, and environment-specific settings.

## Making changes

1. **Infrastructure:** edit `main.go`, run `pulumi preview`.
2. **Lambda logic:** edit `function/handler.py` (linted by Ruff in CI).
3. **Static site:** edit `www/index.html`.
4. Test: `cd pulumi/aws/prod && go test ./...`
5. Deploy: `pulumi up`

## Integration

- Go module `pulumi/aws/prod` in [GitHub Actions CI](../cicd/github-actions-ci.md)
- Python Lambda linted with Ruff
- Separate from Linode [Terraform](terraform.md) / [Packer](packer.md) path

## Official resources

- [Pulumi documentation](https://www.pulumi.com/docs/)
- [Pulumi AWS SDK](https://www.pulumi.com/registry/packages/aws/)
- [Pulumi Go](https://www.pulumi.com/docs/languages-sdks/go/)
