const forms = document.querySelectorAll("form");

forms.forEach((form) => {
  const button = form.querySelector(".submit-button");

  if (!button) {
    console.error("fatal error: could not find submit button for form. define a button with the class 'submit-button' inside form");
  }

  // register submit event for last button found in form
  button.addEventListener("click", (event) => onFormSubmit(event));
});

function onFormSubmit(event: Event) {
  let element = <HTMLElement>event.currentTarget;

  // traverse parent elements from event target to get corresponding form
  while (element.tagName !== "FORM") {
    element = element.parentElement;
  }

  const form = <HTMLFormElement>element;

  if (!form.action) {
    console.error("fatal error: no action defined for this form. cannot parse url for request");
    return;
  }

  fetch(form.action, {
    method: "post",
    body: new FormData(form)
  }).then((response) => {
    // TODO: form validation
    console.log("respose", response);
  });
}
