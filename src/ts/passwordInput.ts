/**
 * This class implements any logic related to password inputs, mainly
 * the 'show password' functionality.
 */
export default class PasswordInput {
  /** CSS class for a 'show password' button */
  private static SHOW_PASSWORD_BUTTON_CLASS = 'show-password-button';

  private constructor() { }

  /**
   * Registers the event listener for every password <input> with 'show password' functionality
   */
  static init() {
    const showPasswordButtons = Array.from(
      document.getElementsByClassName(this.SHOW_PASSWORD_BUTTON_CLASS),
    );

    showPasswordButtons.forEach(
      (button) => {
        const passwordInput = button.parentElement.querySelector('input');
        const icon = button.querySelector('i');

        if (!passwordInput || !icon) {
          throw new Error("error: could not attach 'show password' logic. cannot find corresponding input or i");
        }

        button.addEventListener('click', () => this.togglePasswordVisibility(passwordInput, icon));
      },
    );
  }

  /**
   * Toggles the visibility of a password <input> and sets the corresponding FontAwesome icon
   * @param passwordInput - <input>-Element corresponding to the password
   * @param icon - FontAwesome <i>-Element
   */
  private static togglePasswordVisibility(passwordInput: HTMLInputElement, icon: HTMLElement) {
    const pwInput = passwordInput;
    const ic = icon;

    if (pwInput.type === 'password') {
      pwInput.type = 'text';
      ic.className = 'fa-solid fa-eye fa-fw';
    } else {
      pwInput.type = 'password';
      ic.className = 'fa-solid fa-eye-slash fa-fw';
    }
  }
}
