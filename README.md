# literary-lions
A web forum that allows users to communicate, associate categories with posts, like/dislike posts &amp; comments, and filter posts.

## Person A — Auth & Sessions (me?)

DB schema + migrations for users / sessions (drives the ERD)

Register / login handlers, bcrypt password hashing

Cookie sessions with UUIDs + expiration

Auth middleware (context-based user lookup) exposed as a reusable helper

## Person B — Posts, Comments & Categories
posts / comments / categories tables + at least one CREATE, INSERT, and SELECT query

Create-post (with category) and create-comment handlers

Public read views — posts/comments visible to logged-out users



## Person C — Filtering, Likes and 1 of the suggest bonus functionality (or all if we are fast)
Filter by category / created posts / liked posts (built on A's auth + B's tables)

Shared HTML templates / base layout / CSS

Like/dislike tables + handlers, vote counts shown on display


### dockerfile lets all do it since it's a new thing and everyone should know how to do it.
