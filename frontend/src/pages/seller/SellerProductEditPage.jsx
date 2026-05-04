import { Link, useNavigate, useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import SellerProductForm from '../../components/SellerProductForm.jsx';
import { api } from '../../services/api.js';
import { buttonClass, feedback, layout, pill } from '../../ui/classes.js';

function productToForm(product) {
  return {
    name: product?.name || '',
    description: product?.description || '',
    price: product?.price_cents ? (product.price_cents / 100).toFixed(2) : '',
    commission_percent: product?.commission_percent ?? 30,
  };
}

export default function SellerProductEditPage({ session }) {
  const { id } = useParams();
  const navigate = useNavigate();
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(Boolean(session));
  const [error, setError] = useState('');

  useEffect(() => {
    if (!session || session.user.role !== 'seller') return;

    let active = true;
    setLoading(true);
    setError('');

    api(`/products/${id}`)
      .then((payload) => {
        if (active) setProduct(payload.product);
      })
      .catch((err) => {
        if (active) setError(err.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [id, session]);

  if (!session) {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Sign in to edit products</h1>
        <p className="m-0 text-slate-600">Product editing is available after seller login.</p>
        <Link className={buttonClass('primary')} to="/login">Login</Link>
      </section>
    );
  }

  if (session.user.role !== 'seller') {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Seller account required</h1>
        <p className="m-0 text-slate-600">Only sellers can edit products.</p>
        <Link className={buttonClass('secondary')} to="/products">Back to marketplace</Link>
      </section>
    );
  }

  if (product && product.seller_id !== session.user.id) {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Product unavailable</h1>
        <p className="m-0 text-slate-600">You can only edit products from your own seller catalog.</p>
        <Link className={buttonClass('secondary')} to="/seller/products">Back to products</Link>
      </section>
    );
  }

  const updateProduct = async (payload) => {
    setError('');
    try {
      await api(`/products/${id}`, {
        method: 'PUT',
        body: payload,
      });
      navigate('/seller/products');
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <section className={layout.stack}>
      <div className={layout.pageHeading}>
        <div>
          <p className={layout.eyebrow}>Seller catalog</p>
          <h1 className={layout.title}>Edit product</h1>
          <p className={layout.subtitle}>Update pricing, commission, and product copy for your marketplace listing.</p>
        </div>
        <span className={pill.role}>{session.user.email}</span>
      </div>

      {loading ? <div className={feedback.loading}>Loading product...</div> : null}
      {!loading && !product && error ? <div className={feedback.alert}>{error}</div> : null}
      {product ? (
        <SellerProductForm
          key={product.id}
          initialValues={productToForm(product)}
          submitLabel="Update product"
          submittingLabel="Updating..."
          onSubmit={updateProduct}
          error={error}
        />
      ) : null}
    </section>
  );
}
