import Recaptcha from './recaptcha';
import SessionNotice from './sessionNotice';

/**
 * This class handles all login form related logic, including dynamic input validation.
 */
export default class LoginForm {
  /** login <form> element */
  form: HTMLFormElement;

  /** username <input> element */
  usernameInput: HTMLInputElement;

  /** password <input> element */
  passwordInput: HTMLInputElement;

  /** TOTP <input> element */
  totpInput: HTMLInputElement;

  /** submit <button> element */
  submitButton: HTMLButtonElement;

  private constructor(form: HTMLFormElement) {
    // retrieve mandatory HTMLElements from current form and assign
    this.form = form;
    this.usernameInput = form.querySelector('#inputUsername');
    this.passwordInput = form.querySelector('#inputPassword');
    this.totpInput = form.querySelector('#inputTotp');
    this.submitButton = form.querySelector('button[type="submit"]');

    if (!this.usernameInput || !this.passwordInput || !this.totpInput || !this.submitButton) {
      throw new Error('error: username input, password input, TOTP input or submit button is missing');
    }

    // attach validation logic upon form submission
    form.addEventListener('submit', (event) => this.onFormSubmit(event));
  }

  /** Initialize the given <form> as a login form */
  static init(form: HTMLFormElement): LoginForm {
    return new LoginForm(form);
  }

  /**
   * Dynamically validates the form using Bootstrap form validation.
   * After successful client side validation, the form content will be submitted to the API.
   * Reloads the page after a login, redirecting the user to the actual page.
   * @param event - event that has been triggered by the user
   * @returns void
   */
  async onFormSubmit(event: Event) {
    event.preventDefault();

    const form = <HTMLFormElement>event.currentTarget;

    // validate form before sending POST request
    if (form.classList.contains('needs-validation')) {
      form.classList.add('was-validated');

      if (!form.checkValidity()) {
        return;
      }
    }

    if (!form.action) {
      console.error('fatal error: no action defined for this form. cannot parse url for request');
      return;
    }

    // save form data
    const formDataObject = {};
    const formData = new FormData(form);
    formData.forEach((value, key) => {
      formDataObject[key] = value;
    });

    // replace button text with a spinner while the request is ongoing
    const originalButtonHTML = this.submitButton.innerHTML;
    const buttonWidth = this.submitButton.getBoundingClientRect().width;

    this.submitButton.setAttribute('style', `width: ${buttonWidth}px;`);
    this.submitButton.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i>';

    this.toggleFormState();

    let recaptchaToken = '';

    // execute reCAPTCHA if it's enabled
    if (Recaptcha.ENABLED) {
      recaptchaToken = await Recaptcha.execute();
    }

    // submit the form to the API for server-side validation
    try {
      const response = await fetch(form.action, {
        method: 'post',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...formDataObject,
          recaptchaToken, // attach reCAPTCHA token to the request
        }),
      });

      if (response.ok) {
        // get cookie expiration timestamp from response body and save it to localStorage
        response.json()
          .then((json) => {
            localStorage.setItem(
              SessionNotice.TOKEN_EXPIRATION_LOCALSTORAGE_KEY,
              String(json.expires),
            );
          })
          .catch((error) => {
            console.error('error: could not convert response to json.', error);
          });

        // reload the page if the API reports a successful login
        window.location.reload();
      } else {
        this.resetSubmitButton(originalButtonHTML);
        this.toggleFormState();

        // process API response to determine error origin
        const responseText = await response.text();

        if (responseText.includes('TOTP')) {
          this.usernameInput.readOnly = true;
          this.passwordInput.readOnly = true;

          this.totpInput.setCustomValidity('Invalid TOTP.');
          this.submitButton.disabled = true;

          // clear error message after value change on TOTP input
          this.totpInput.addEventListener('input', () => {
            this.totpInput.setCustomValidity('');
            this.submitButton.disabled = false;
          }, { once: true });

          this.totpInput.parentElement.classList.remove('d-none');
        } else {
          this.usernameInput.setCustomValidity('Invalid credentials.');
          this.passwordInput.setCustomValidity('Invalid credentials.');
          this.submitButton.disabled = true;

          // clear error messages after value changes on either input
          this.usernameInput.addEventListener('input', () => {
            this.usernameInput.setCustomValidity('');
            this.passwordInput.setCustomValidity('');
            this.submitButton.disabled = false;
          }, { once: true });

          this.passwordInput.addEventListener('input', () => {
            this.usernameInput.setCustomValidity('');
            this.passwordInput.setCustomValidity('');
            this.submitButton.disabled = false;
          }, { once: true });
        }
      }
    } catch (error) {
      this.resetSubmitButton(originalButtonHTML);
      this.toggleFormState();
      console.error(error);

      // TODO: better error-handling when the server is offline for some reason
    }
  }

  /** Resets the submit button to the initial state. */
  resetSubmitButton(originalHtml: string): void {
    this.submitButton.innerHTML = originalHtml;
    this.submitButton.removeAttribute('style');
  }

  /** Disables/enables form, it's inputs and buttons */
  toggleFormState(): void {
    this.usernameInput.readOnly = !this.usernameInput.readOnly;
    this.passwordInput.readOnly = !this.passwordInput.readOnly;
    this.submitButton.disabled = !this.submitButton.disabled;
  }
}
