import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../../services/api.js';
import { saveSession } from '../../services/session.js';
import { buttonClass, feedback, form as formClasses, layout, panel } from '../../ui/classes.js';

export default function RegisterPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: '', password: '', role: 'affiliate' });
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
      const authResponse = await api('/auth/register', {
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
        <p className={layout.eyebrow}>Marketplace access</p>
        <h1 className="mb-4 text-4xl font-extrabold leading-tight text-slate-950">Create account</h1>
        <p className="max-w-xl text-base text-slate-600">Choose the role that matches how you will use the platform.</p>
      </div>

      <form className={`${panel.base} ${formClasses.root} p-6`} onSubmit={submit}>
        <div>
          <h2 className="mb-2 text-lg font-extrabold leading-snug text-slate-950">Register</h2>
          <p className={layout.muted}>Affiliate is the default role.</p>
        </div>

        {error ? <div className={feedback.alert}>{error}</div> : null}

        <label className={formClasses.label}>
          Email
          <input className={formClasses.control} type="email" name="email" autoComplete="email" value={form.email} onChange={update} />
        </label>
        <label className={formClasses.label}>
          Password
          <input className={formClasses.control} type="password" name="password" autoComplete="new-password" value={form.password} onChange={update} />
        </label>
        <label className={formClasses.label}>
          Role
          <select className={formClasses.control} name="role" value={form.role} onChange={update}>
            <option value="affiliate">Affiliate</option>
            <option value="seller">Seller</option>
          </select>
        </label>
        <button className={buttonClass('primary', 'w-full')} type="submit" disabled={loading}>
          {loading ? 'Creating account...' : 'Register'}
        </button>
        <p className={formClasses.footer}>
          Already registered? <Link to="/login">Login</Link>
        </p>
      </form>
    </section>
  );
}
