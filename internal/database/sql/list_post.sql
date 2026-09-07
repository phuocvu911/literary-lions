   SELECT
        p.id,
        p.title,
        p.content,
        p.created_at,
        u.username,
        (
            SELECT COUNT(*)
            FROM comments AS c
            WHERE c.post_id = p.id
        ) AS comment_count,
        (
            SELECT COUNT(*)
            FROM post_reactions AS pr
            WHERE pr.post_id = p.id
              AND pr.value = 1
        ) AS likes,
        (
            SELECT COUNT(*)
            FROM post_reactions AS pr
            WHERE pr.post_id = p.id
              AND pr.value = -1
        ) AS dislikes
        COALESCE(
             (
                SELECT GROUP_CONCAT(c.name, ', ')
                FROM post_categories AS pc
                INNER JOIN categories AS c
                    ON c.id = pc.category_id
                WHERE pc.post_id = p.id
            ),
            ''
        ) AS category_names
    FROM posts AS p
    INNER JOIN users AS u
        ON u.id = p.user_id
    ORDER BY p.created_at DESC
