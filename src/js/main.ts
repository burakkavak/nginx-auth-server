class LoginForm {
  form: HTMLFormElement;
  usernameInput: HTMLInputElement;
  passwordInput: HTMLInputElement;
  totpInput: HTMLInputElement;
  submitButton: HTMLButtonElement;

  constructor(form: HTMLFormElement) {
    this.form = form;
    this.usernameInput = form.querySelector("#inputUsername");
    this.passwordInput = form.querySelector("#inputPassword");
    this.totpInput = form.querySelector("#inputTotp");
    this.submitButton = form.querySelector('button[type="submit"]');

    if (!this.usernameInput || !this.passwordInput || !this.totpInput || !this.submitButton) {
      throw new Error("error: username input, password input, TOTP input or submit button is missing");
    }

    form.addEventListener("submit", (event) => this.onFormSubmit(event))
  }

  async onFormSubmit(event: Event) {
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

    const formData = new FormData(form);

    // replace button text with a spinner while the request is ongoing
    const originalButtonHTML = this.submitButton.innerHTML;
    const buttonWidth = this.submitButton.getBoundingClientRect().width;

    this.submitButton.setAttribute("style", `width: ${buttonWidth}px;`);
    this.submitButton.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i>';

    this.toggleFormState();

    try {
      const response = await fetch(form.action, {
        method: "post",
        body: formData
      });

      if (response.ok) {
        location.reload()
      } else {
        this.resetSubmitButton(originalButtonHTML);
        this.toggleFormState();

        const responseText = await response.text();

        if (responseText.includes("TOTP")) {
          this.usernameInput.readOnly = true;
          this.passwordInput.readOnly = true;

          this.totpInput.setCustomValidity("Invalid TOTP.");
          this.submitButton.disabled = true;

          // clear error message after value change on TOTP input
          this.totpInput.addEventListener("input", () => {
            this.totpInput.setCustomValidity("");
            this.submitButton.disabled = false;
          }, {once: true});

          this.totpInput.parentElement.classList.remove("d-none");
        } else {
          this.usernameInput.setCustomValidity("Invalid credentials.");
          this.passwordInput.setCustomValidity("Invalid credentials.");
          this.submitButton.disabled = true;

          // clear error messages after value changes on either input
          this.usernameInput.addEventListener("input", () => {
            this.usernameInput.setCustomValidity("");
            this.passwordInput.setCustomValidity("");
            this.submitButton.disabled = false;
          }, {once: true});

          this.passwordInput.addEventListener("input", () => {
            this.usernameInput.setCustomValidity("");
            this.passwordInput.setCustomValidity("");
            this.submitButton.disabled = false;
          }, {once: true});
        }
      }
    } catch (error) {
      this.resetSubmitButton(originalButtonHTML);
      this.toggleFormState();
      console.error(error);

      // TODO: better error-handling when the server is offline for some reason
    }
  }

  resetSubmitButton(originalHtml: string): void {
    this.submitButton.innerHTML = originalHtml;
    this.submitButton.removeAttribute("style");
  }

  /**
   * Disables/enables form, inputs and buttons
   */
  toggleFormState(): void {
    this.usernameInput.readOnly = !this.usernameInput.readOnly
    this.passwordInput.readOnly = !this.passwordInput.readOnly
    this.submitButton.disabled = !this.submitButton.disabled
  }
}

const loginForm = <HTMLFormElement>document.querySelector("form.login-form");

if (loginForm) {
  new LoginForm(loginForm)
}
