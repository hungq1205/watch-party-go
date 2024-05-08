let isCreate = false;
const title = document.getElementById("operation-title");
const alterOption = document.getElementById("alter-option");
const boxId = document.getElementById("box-id");
const submit = document.getElementById("submit");
const form = document.querySelector("form");

window.onload = function() {
    alterOption.onclick = function() {
        isCreate = !isCreate;
        createForm(isCreate);
    };
};

form.addEventListener("submit", e => {
    e.preventDefault();
    const formData = new FormData(form);
    const data = new URLSearchParams(formData);

    if (isCreate) {
        data.delete("box_id");
        fetch('http://localhost:3000/create', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: data
        }).then(function(res) {
            if (res.status == 201 || res.status == 200) 
                {
                    alert("Created");
                    res.json().then( function(jdata) {
                        data.append("box_id", jdata["BoxId"])
                        joinBox(data)
                    });
                }
            else
                res.text()
                    .then(text => alert(text))
                    .catch(err => alert(err));
        })
        .catch(err => alert(err))
    }
    else {
        joinBox(data)
    }
});

function joinBox(fdata) {
    fetch('http://localhost:3000/join', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: fdata
    }).then(function(res) {
        if (res.status == 200) {
            window.location.href = "http://localhost:3000/box";
        }
        else {
            res.text()
            .then(text => alert(text))
            .catch(err => alert(err));
        }
    })
    .catch(err => alert(err));
}

function createForm(isActive)
{
    if (isActive)
    {
        title.textContent = "Create Box"
        alterOption.textContent = "Join";
        boxId.parentElement.classList.add("d-none");
        submit.textContent = "Create";
    }
    else
    {
        title.textContent = "Join Box"
        alterOption.textContent = "Create";
        boxId.parentElement.classList.remove("d-none");
        submit.textContent = "Join";
    }
}