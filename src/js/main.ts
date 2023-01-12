import LoginForm from './loginForm';
import Recaptcha from './recaptcha';
import SessionNotice from './sessionNotice';
import PasswordInput from './passwordInput';

// Initialize Google reCAPTCHA if the container is set in the template
if (document.querySelector('#g-recaptcha')) {
  Recaptcha.init();
}

const sessionNoticeContainer = document.getElementById('sessionExpiredNotice');

if (sessionNoticeContainer) {
  SessionNotice.init(sessionNoticeContainer);
}

const loginForm = <HTMLFormElement>document.querySelector('form.login-form');

// Initialize login logic if the login form is present
if (loginForm) {
  LoginForm.init(loginForm);
}

PasswordInput.init();
