## Proposal: Add Costory team management resources

### Goal
Extend the Terraform provider to manage Costory Teams and team members using the
`/terraform/teams` API endpoints.

### Planned changes
- Add Costory API client support for Teams (create/read/update/archive) and
  team member mutations (add/remove).
- Introduce a `costory_team` resource with `name`, `description`, `visibility`
  and computed `id`, `created_at`, `updated_at` attributes.
- Introduce a `costory_team_member` resource to add/remove a single member,
  supporting either `user_id` or `email` plus optional `role`.
- Register the new resources in the provider.
- Regenerate docs under `docs/`.

### Validation
- Run `go test ./...`.
- Run `scripts/generate-docs.sh` if schemas change.
