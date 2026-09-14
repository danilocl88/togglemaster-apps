# ToggleMaster Apps

Monorepo dos microservicos utilizados na Fase 3 do Tech Challenge.

## Services

| Service | Language | Port |
|---|---|---:|
| auth-service | Go | 8001 |
| flag-service | Python | 8002 |
| targeting-service | Python | 8003 |
| evaluation-service | Go | 8004 |
| analytics-service | Python | 8005 |

## Architecture

The application is composed of five independently deployable
microservices.

Runtime infrastructure is maintained separately in:

- `togglemaster-infra` - AWS/Terraform infrastructure
- `togglemaster-gitops` - Kubernetes/GitOps manifests

## Container Security

Each service uses its own Dockerfile and runs using a non-root
application user.

Secrets and runtime credentials are not stored in this repository.

## CI/CD

Independent GitHub Actions workflows for each microservice will provide:

- build and tests;
- lint;
- SAST/SCA;
- container build;
- Trivy image scanning;
- ECR publishing using immutable commit-based tags;
- GitOps manifest updates.
