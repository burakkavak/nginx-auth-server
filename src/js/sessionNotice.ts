/**
 * This class handles the re-authentication notice. Upon expiration of the authentication cookie,
 * the user has to authenticate again. The user will then will be greeted by a notice that the
 * previous session has expired.
 * The cookie expiration date is set by the login form upon successful authentication and saved
 * to the browsers localStorage.
 */
export default class SessionNotice {
  static TOKEN_EXPIRATION_LOCALSTORAGE_KEY = 'tokenExpiration';

  private constructor(noticeContainer: HTMLElement) {
    const expires = localStorage.getItem(SessionNotice.TOKEN_EXPIRATION_LOCALSTORAGE_KEY);

    if (expires && expires !== '' && Date.now() > +expires) {
      noticeContainer.classList.remove('d-none');
      localStorage.removeItem(SessionNotice.TOKEN_EXPIRATION_LOCALSTORAGE_KEY);
    }
  }

  static init(noticeContainer: HTMLElement): SessionNotice {
    return new SessionNotice(noticeContainer);
  }
}
