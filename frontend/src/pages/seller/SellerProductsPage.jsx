import { Link } from 'react-router-dom';
import { useCallback, useEffect, useState } from 'react';
import SellerProductManager from '../../components/SellerProductManager.jsx';
import { api } from '../../services/api.js';
import { buttonClass, feedback, layout, pill } from '../../ui/classes.js';

export default function SellerProductsPage({ session }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(Boolean(session));
  const [error, setError] = useState('');

  const loadProducts = useCallback(async () => {
    if (!session || session.user.role !== 'seller') return;

    setLoading(true);
    setError('');
    try {
      const payload = await api('/dashboard/seller');
      setData(payload);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [session]);

  useEffect(() => {
    loadProducts();
  }, [loadProducts]);

  if (!session) {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Sign in to manage products</h1>
        <p className="m-0 text-slate-600">Seller products are available after login.</p>
        <Link className={buttonClass('primary')} to="/login">Login</Link>
      </section>
    );
  }

  if (session.user.role !== 'seller') {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Seller account required</h1>
        <p className="m-0 text-slate-600">Only sellers can manage product catalogs.</p>
        <Link className={buttonClass('secondary')} to="/products">Back to marketplace</Link>
      </section>
    );
  }

  return (
    <section className={layout.stack}>
      <div className={layout.pageHeading}>
        <div>
          <p className={layout.eyebrow}>Seller catalog</p>
          <h1 className={layout.title}>Products</h1>
          <p className={layout.subtitle}>Create and maintain the products customers see in the public marketplace.</p>
        </div>
        <span className={pill.role}>{session.user.email}</span>
      </div>

      {loading ? <div className={feedback.loading}>Loading products...</div> : null}
      {error ? <div className={feedback.alert}>{error}</div> : null}
      {data ? <SellerProductManager products={data.products || []} onChanged={loadProducts} /> : null}
    </section>
  );
}
