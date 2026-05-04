import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../../services/api.js';
import { saveSession } from '../../services/session.js';
import { buttonClass, feedback, form as formClasses, layout, panel } from '../../ui/classes.js';

export default function LoginPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: '', password: '' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const update = (event) => {
    setForm((current) => ({ ...current, [event.target.name]: event.target.value }));
  };

  const submit = async (event) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const authResponse = await api('/auth/login', {
        method: 'POST',
        body: form,
      });
      saveSession(authResponse);
      navigate('/dashboard');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="grid min-h-[calc(100vh-180px)] grid-cols-1 items-center gap-8 md:grid-cols-[minmax(0,1fr)_minmax(320px,430px)] md:gap-16 lg:gap-24">
      <div>
        <p className={layout.eyebrow}>Affiliate operations</p>
        <h1 className="mb-4 text-4xl font-extrabold leading-tight text-slate-950">Welcome back</h1>
        <p className="max-w-xl text-base text-slate-600">Track commissions, manage products, and keep revenue attribution clean from one workspace.</p>
      </div>

      <form className={`${panel.base} ${formClasses.root} p-6`} onSubmit={submit}>
        <div>
          <h2 className="mb-2 text-lg font-extrabold leading-snug text-slate-950">Login</h2>
          <p className={layout.muted}>Use your marketplace account.</p>
        </div>

        {error ? <div className={feedback.alert}>{error}</div> : null}

        <label className={formClasses.label}>
          Email
          <input className={formClasses.control} type="email" name="email" autoComplete="email" value={form.email} onChange={update} />
        </label>
        <label className={formClasses.label}>
          Password
          <input className={formClasses.control} type="password" name="password" autoComplete="current-password" value={form.password} onChange={update} />
        </label>
        <button className={buttonClass('primary', 'w-full')} type="submit" disabled={loading}>
          {loading ? 'Signing in...' : 'Login'}
        </button>
        <p className={formClasses.footer}>
          No account yet? <Link to="/register">Register</Link>
        </p>
      </form>
    </section>
  );
}
