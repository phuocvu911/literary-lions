//for icons
lucide.createIcons();

document.getElementById("profile-image").addEventListener("change", function (e) {
    if (this.files.length > 0) {
        document.getElementById("profile-upload").submit();
    }
});

const editBtn = document.getElementById("edit-profile");
const modal = document.getElementById("edit-modal");
const closeBtn = document.querySelector(".close");

editBtn.addEventListener("click", () => {
    modal.style.display = "block";
});

closeBtn.addEventListener("click", () => {
    modal.style.display = "none";
});

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