import { firstValueFrom, Subject } from 'rxjs';

/**
 * This class handles all Google reCAPTCHA related logic.
 */
export default class Recaptcha {
  /**
   * True if reCAPTCHA is enabled.
   * Used by {@link LoginForm}
   */
  public static ENABLED = false;

  /** On a successful Captcha, the Google-provided tokens are saved here  */
  private static TOKEN$ = new Subject<string>();

  private constructor() { }

  static init() {
    this.ENABLED = true;
  }

  /**
   * Called after reCAPTCHA loaded it's scripts.
   * Will render the widget on the current site.
   */
  static onLoad() {
    const container = <HTMLElement>document.querySelector('#g-recaptcha');

    if (!container) {
      throw new Error('error: reCAPTCHA container does not exist. cannot render widget.');
    }

    grecaptcha.render(container, {
      callback: (response) => this.onExecute(response),
    }, true);
  }

  /**
   * Executes the captcha, retrieving a token from Google upon successful captcha solving.
   * Executed by {@link LoginForm}
   * @returns reCAPTCHA token that shall be verified server-side
   */
  static async execute(): Promise<string> {
    grecaptcha.execute();

    const token = await firstValueFrom(this.TOKEN$);

    return token;
  }

  /**
   * Called by reCAPTCHA.js after the captcha has been solved.
   * @param response - reCAPTCHA token that shall be verified server-side
   */
  static onExecute(response: string) {
    grecaptcha.reset();

    this.TOKEN$.next(response);
  }
}

// redirect window.recaptchaOnLoad to this class
(<any>window).recaptchaOnLoad = () => Recaptcha.onLoad();
