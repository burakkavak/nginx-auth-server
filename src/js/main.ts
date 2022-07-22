import LoginForm from './loginForm';
import Recaptcha from './recaptcha';

if (document.querySelector('#g-recaptcha')) {
  Recaptcha.init();
}

const loginForm = <HTMLFormElement>document.querySelector('form.login-form');

if (loginForm) {
  LoginForm.init(loginForm);
}
