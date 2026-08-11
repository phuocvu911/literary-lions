# Literary Lions Forum — Task Split
 
A Go web forum for a book club: SQLite-backed, cookie/UUID sessions, categorized
posts, like/dislike, filtering, and Docker deployment.
 
## Before splitting
 
Agree together, in one sitting, before writing code:
 
- **ERD / DB schema** — tables, columns, foreign keys for `users`, `sessions`,
  `posts`, `comments`, `categories`, `post_categories`, `likes`.
- **Route list** — every HTTP endpoint and method, e.g. `POST /register`,
  `GET /posts?category=`, `POST /posts/{id}/like`.
- **Migration file naming** — e.g. `001_users.sql`, `002_posts.sql`, so schema
  changes from different people don't collide.
- **Auth middleware contract** — function signature Person A will hand to B
  and C (e.g. how the logged-in user is retrieved in a handler), agreed up
  front so B/C can stub it and start immediately.
- **Store package shape** — DB access behind an interface (e.g. `store`
  package) so handlers can be unit-tested with `httptest` against a fake
  store before the real DB logic lands.
- **Base template / shared CSS** — a minimal shared layout (nav, base
  styles) so the three people's pages don't look like three different
  websites glued together.
## Roles
 
Each person owns one backend domain end-to-end — DB, handlers, own
templates/CSS for their pages, and their own Dockerfile.
 
### Person A — Auth & Sessions (me?)
*Critical path: B and C's handlers depend on this, so ship a stub early.*
 
- DB schema + migrations for `users` / `sessions` (drives the ERD)
- Register / login handlers, bcrypt password hashing
- Cookie sessions with UUIDs + expiration
- Auth middleware (context-based user lookup) exposed as a reusable helper
- Templates/CSS for: register, login, profile/nav auth state
- Own Dockerfile (see Docker below)
### Person B — Posts, Comments & Categories
 
- `posts` / `comments` / `categories` tables + at least one CREATE, INSERT,
  and SELECT query
- Create-post (with category) and create-comment handlers
- Public read views — posts/comments visible to logged-out users
- Templates/CSS for: post list, single post + comments, new-post form
- A search feature
- Own Dockerfile (see Docker below)
### Person C — Likes & Filtering
 
- Like/dislike tables + handlers, vote counts shown on display
- Filter by category / created posts / liked posts (built on A's auth +
  B's tables)
- Templates/CSS for: like/dislike UI, filter controls
- Allow users to upload images or other files
- User profile page
- Own Dockerfile (see Docker below)
## Docker — everyone writes their own
 
Since Docker is new to all three of you, each person writes their own
Dockerfile against the shared app (not just their slice) as a learning
exercise, rather than one person doing it for the team.
 
- Each person builds and runs the full app in their own container,
  independently, once the app compiles.
- Compare the three Dockerfiles as a group afterward — discuss what
  differs (base image, layer caching, multi-stage build choices,
  labels/metadata) and merge into a single team Dockerfile using the
  best ideas from each.
- Whoever's version (or hybrid) is adopted still needs the "clean
  environment" requirement covered: pruning unused images/containers.
