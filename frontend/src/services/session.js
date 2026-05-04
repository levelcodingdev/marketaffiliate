export function getSession() {
  const token = window.localStorage.getItem('auth_token');
  const userJSON = window.localStorage.getItem('auth_user');
  if (!token || !userJSON) return null;

  try {
    return {
      token,
      user: JSON.parse(userJSON),
    };
  } catch {
    clearSession();
    return null;
  }
}

export function saveSession(authResponse) {
  window.localStorage.setItem('auth_token', authResponse.token);
  window.localStorage.setItem('auth_user', JSON.stringify(authResponse.user));
  window.dispatchEvent(new Event('session-change'));
}

export function clearSession() {
  window.localStorage.removeItem('auth_token');
  window.localStorage.removeItem('auth_user');
  window.dispatchEvent(new Event('session-change'));
}
