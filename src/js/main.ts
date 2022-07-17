const forms = document.querySelectorAll("form");

forms.forEach((form) => {
  const button = form.querySelector(".submit-button");

  if (!button) {
    console.error("fatal error: could not find submit button for form. define a button with the class 'submit-button' inside form");
  }

  form.addEventListener("submit", (event) => onFormSubmit(event))
});

function onFormSubmit(event: Event) {
  event.preventDefault();

  const form = <HTMLFormElement>event.currentTarget;

  // validate form before sending POST request
  if (form.classList.contains("needs-validation")) {
    form.classList.add('was-validated')

    if (!form.checkValidity()) {
      return;
    }
  }

  if (!form.action) {
    console.error("fatal error: no action defined for this form. cannot parse url for request");
    return;
  }

  fetch(form.action, {
    method: "post",
    body: new FormData(form)
  }).then((response) => {
    if (response.ok) {
      location.reload()
    } else {
      // TODO: form validation
    }
  });
}
