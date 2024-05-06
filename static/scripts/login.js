let isSignup = false;
const title = document.getElementById("operation-title");
const alterOption = document.getElementById("alter-option");
const displayName = document.getElementById("display-name");
const submit = document.getElementById("submit");
const form = document.querySelector("form");

window.onload = function() {
    alterOption.onclick = function() {
        isSignup = !isSignup;
        signUpForm(isSignup);
    };
};

form.addEventListener("submit", e => {
    e.preventDefault();
    const formData = new FormData(form);
    const data = new URLSearchParams(formData);

    data.set("username", data.get("username").trim());

    if (isSignup) {
        data.set("display_name", data.get("display_name").trim());
        fetch('http://localhost:3000/signup', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: data
        }).then(function(res) {
            if (res.status == 201) 
                alert("Signed up");
            else
                res.text()
                    .then(text => alert(text))
                    .catch(err => alert(err));
        })
        .catch(err => alert(err))
    }
    else {
        data.delete("display_name");
        fetch('http://localhost:3000/login', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: data
        }).then(res => res.text())
            .then(text => alert(text))
            .catch(err => alert(err));
    }
});

function signUpForm(isActive)
{
    if (isActive)
    {
        title.textContent = "Sign Up"
        alterOption.textContent = "Log in";
        displayName.parentElement.classList.remove("d-none");
        submit.textContent = "Sign up";
    }
    else
    {
        title.textContent = "Please log in"
        alterOption.textContent = "Sign up";
        displayName.parentElement.classList.add("d-none");
        submit.textContent = "Log in";
    }
}