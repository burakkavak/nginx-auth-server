import LoginForm from './loginForm';
import Recaptcha from './recaptcha';

// Initialize Google reCAPTCHA if the container is set in the template
if (document.querySelector('#g-recaptcha')) {
  Recaptcha.init();
}

const loginForm = <HTMLFormElement>document.querySelector('form.login-form');

// Initialize login logic if the login form is present
if (loginForm) {
  LoginForm.init(loginForm);
}
