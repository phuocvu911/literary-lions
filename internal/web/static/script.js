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

const editBtn = document.getElementById("edit-profile");
const modal = document.getElementById("edit-modal");
const closeBtn = document.querySelector(".close");

if (editBtn) {
    editBtn.addEventListener("click", () => {
        modal.style.display = "block";
    });
}

if (closeBtn) {
    closeBtn.addEventListener("click", () => {
        modal.style.display = "none";
    });
}


window.addEventListener("click", (e) => {
    if (e.target === modal) {
        modal.style.display = "none";
    }
});      

function validateForm() {
    const username = document.forms["edit-form"]["username"].value;

    if (!/^[^\s@]{3,100}$/.test(username)) {
        displayError("Username must be at least 3 characters, with no spaces or '@'.", "username");
        return false;
    }

    const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^A-Za-z\d]).{8,100}$/;
    const password = document.forms["edit-form"]["password"].value;

    if (!passwordRegex.test(password)) {
        displayError("Password must be at least 8 characters and contain uppercase, lowercase, number, and special character.", "password");
        return false;
    }

    return true;
}

function displayError(msg, type) {
    const oldError = document.querySelector(".form-error");
    if (oldError) {
        oldError.remove();
    }

    const username = document.getElementById("username");
    const password = document.getElementById("password");

    const div = document.createElement("div");
    div.classList.add("form-error");
    div.innerText = msg;            

    if (type === "username") {
        username.parentElement.appendChild(div);
    } else if (type === "password") {
        password.parentElement.appendChild(div);
    }
    
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