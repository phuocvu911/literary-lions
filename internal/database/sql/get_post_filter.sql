SELECT
    posts.id,
    posts.title,
    posts.content,
    posts.created_at,
    users.username AS author,

    COUNT(DISTINCT comments.id) AS comment_count,

    COUNT(DISTINCT CASE
        WHEN post_reactions.value = 1
        THEN post_reactions.user_id
    END) AS likes,

    COUNT(DISTINCT CASE
        WHEN post_reactions.value = -1
        THEN post_reactions.user_id
    END) AS dislikes,

    GROUP_CONCAT(DISTINCT categories.name) AS category_names

FROM posts

JOIN users
    ON users.id = posts.user_id

LEFT JOIN comments
    ON comments.post_id = posts.id

LEFT JOIN post_reactions
    ON post_reactions.post_id = posts.id

LEFT JOIN post_categories
    ON post_categories.post_id = posts.id

LEFT JOIN categories
    ON categories.id = post_categories.category_id