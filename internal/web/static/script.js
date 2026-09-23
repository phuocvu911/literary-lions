//for icons
lucide.createIcons();

const profileImage = document.getElementById("profile-image");

if (profileImage) {
    profileImage.addEventListener("change", function () {
        if (this.files.length > 0) {
            document.getElementById("profile-upload").submit();
        }
    });
}

const deleteProfile = document.getElementById('delete-profile')

if (deleteProfile) {
    deleteProfile.addEventListener('click', async function () {
        if (!confirm('Are you sure you want to delete your profile?')) {
            return;
        }

        try {
            const response = await fetch('/profile/delete', {
                method: 'DELETE'
            });

            if (response.ok) {
                window.location.href = '/';
            } else {
                alert('Failed to delete profile.');
            }
        } catch (error) {
            console.error(error);
            alert('Something went wrong.');
        }
    });
}

const reactionsOnPost = document.querySelectorAll(".reaction-buttons-posts");

if (reactionsOnPost) {
    async function react(value, postID, likeCount, dislikeCount, likeBtn, dislikeBtn) {
        const response = await fetch(`/post/${postID}/reaction`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ value })
        });

        if (!response.ok) {
            console.error("Failed to update reaction");
            const error = await response.text();
            console.error("Reaction failed:", response.status, error);
            return;
        }

        const data = await response.json();

        likeCount.textContent = data.likes;
        dislikeCount.textContent = data.dislikes;

        likeBtn.classList.toggle("active", data.reaction_to_post === 1);
        dislikeBtn.classList.toggle("active", data.reaction_to_post === -1);
    }

    reactionsOnPost.forEach((reaction) => {
        const postID = reaction.dataset.postId;
        const likeCount = reaction.querySelector("#likePostCount");
        const dislikeCount = reaction.querySelector("#dislikePostCount");
        const likeBtn = reaction.querySelector("#likePostBtn");
        const dislikeBtn = reaction.querySelector("#dislikePostBtn");

        if (likeBtn) {
            likeBtn.addEventListener("click", () => {
                react(1, postID, likeCount, dislikeCount, likeBtn, dislikeBtn);
            });
        }

        if (dislikeBtn) {
            dislikeBtn.addEventListener("click", () => {
                react(-1, postID, likeCount, dislikeCount, likeBtn, dislikeBtn);
            });
        }


    });  
}

const reactionsOnComment = document.querySelectorAll(".reaction-buttons-comments");

if (reactionsOnComment) {
    async function react(value, commentID, likeCount, dislikeCount, likeBtn, dislikeBtn) {
        const response = await fetch(`/comment/${commentID}/reaction`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ value })
        });

        if (!response.ok) {
            console.error("Failed to update reaction");
            const error = await response.text();
            console.error("Reaction failed:", response.status, error);
            return;
        }

        const data = await response.json();

        likeCount.textContent = data.likes;
        dislikeCount.textContent = data.dislikes;

        likeBtn.classList.toggle("active", data.reaction_to_comment === 1);
        dislikeBtn.classList.toggle("active", data.reaction_to_comment === -1);
    }

    reactionsOnComment.forEach((reaction) => {
        const commentID = reaction.dataset.commentId;
        const likeCount = reaction.querySelector("#likeCommentCount");
        const dislikeCount = reaction.querySelector("#dislikeCommentCount");
        const likeBtn = reaction.querySelector("#likeCommentBtn");
        const dislikeBtn = reaction.querySelector("#dislikeCommentBtn");

        if (likeBtn) {
            likeBtn.addEventListener("click", () => {
                react(1, commentID, likeCount, dislikeCount, likeBtn, dislikeBtn);
            });
        }

        if (dislikeBtn) {
            dislikeBtn.addEventListener("click", () => {
                react(-1, commentID, likeCount, dislikeCount, likeBtn, dislikeBtn);
            });
        }


    });
}