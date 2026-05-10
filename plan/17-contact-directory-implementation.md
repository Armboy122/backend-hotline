# HNX-01 Contact Directory Backend Implementation

Date: 2026-05-10
Status: implemented in backend

## Scope

Implemented the contact directory backend support from the HNP-01 contract:

- `GET /v1/contact-directory` lists active users with contact fields and supports search/filter query params.
- `GET /v1/contact-directory/:userId` returns one contact entry.
- `PATCH /v1/users/me/contact` lets an authenticated user update their own display name, position, and phone number.
- `PATCH /v1/users/:id/contact` lets `super_admin` update another user's contact fields.

## Query parameters

`GET /v1/contact-directory` accepts:

| Param | Notes |
|---|---|
| `query` | Searches username, display name, position, and phone number. |
| `teamId` | Optional team filter. |
| `role` | Optional role filter. |
| `includeInactive` | Honored only for `super_admin`; other roles are forced to active contacts only. |
| `page` / `limit` | Defaults to `1` / `50`; max limit is `100`. |

## Contact fields

The `User` model now has nullable contact fields:

- `displayName`
- `position`
- `phoneNumber`

These are additive nullable columns and keep existing `username` authentication semantics unchanged. The current project uses GORM `AutoMigrate`, so these fields are added through the existing migration path rather than a standalone SQL migration file.

## RBAC

- Every authenticated role can view active contact directory entries.
- Non-`super_admin` callers cannot list inactive users even if `includeInactive=true` is supplied.
- A user can update only their own contact fields.
- `super_admin` can update any user's contact fields.
- `admin` is not treated as a broad user manager for contact edits, matching the HNP-01/HNP-00 RBAC split.

## Verification

Required backend gates for this card:

```bash
go test ./...
go vet ./...
```
