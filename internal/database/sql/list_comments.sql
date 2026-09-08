SELECT
    c.id,
    c.content,
    c.created_at,
    u.username,
    (
        SELECT COUNT(*)
        FROM comment_reactions AS cr
        WHERE cr.comment_id = c.id
          AND cr.value = 1
    ) AS likes,
    (
        SELECT COUNT(*)
        FROM comment_reactions AS cr
        WHERE cr.comment_id = c.id
          AND cr.value = -1
    ) AS dislikes
FROM comments AS c
INNER JOIN users AS u
    ON u.id = c.user_id
WHERE c.post_id = ?
ORDER BY c.created_at ASC