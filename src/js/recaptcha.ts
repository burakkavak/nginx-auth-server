import { firstValueFrom, Subject } from 'rxjs';

export default class Recaptcha {
  public static ENABLED = false;

  private static TOKEN$ = new Subject<string>();

  private constructor() { }

  static init() {
    this.ENABLED = true;
  }

  static onLoad() {
    const container = <HTMLElement>document.querySelector('#g-recaptcha');

    if (!container) {
      throw new Error('error: reCAPTCHA container does not exist. cannot render widget.');
    }

    grecaptcha.render(container, {
      callback: (response) => this.onExecute(response),
    }, true);
  }

  static async execute(): Promise<string> {
    grecaptcha.execute();

    const token = await firstValueFrom(this.TOKEN$);

    return token;
  }

  static onExecute(response: string) {
    grecaptcha.reset();

    this.TOKEN$.next(response);
  }
}

(<any>window).recaptchaOnLoad = () => Recaptcha.onLoad();
