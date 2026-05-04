import { Link, useNavigate } from 'react-router-dom';
import SellerProductForm from '../../components/SellerProductForm.jsx';
import { api } from '../../services/api.js';
import { buttonClass, feedback, layout, pill } from '../../ui/classes.js';
import { useState } from 'react';

export default function SellerProductCreatePage({ session }) {
  const navigate = useNavigate();
  const [error, setError] = useState('');

  if (!session) {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Sign in to create products</h1>
        <p className="m-0 text-slate-600">Product creation is available after seller login.</p>
        <Link className={buttonClass('primary')} to="/login">Login</Link>
      </section>
    );
  }

  if (session.user.role !== 'seller') {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Seller account required</h1>
        <p className="m-0 text-slate-600">Only sellers can create products.</p>
        <Link className={buttonClass('secondary')} to="/products">Back to marketplace</Link>
      </section>
    );
  }

  const createProduct = async (payload) => {
    setError('');
    try {
      await api('/products', {
        method: 'POST',
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
          <h1 className={layout.title}>Create product</h1>
          <p className={layout.subtitle}>Add a product that customers can browse and buy in the public marketplace.</p>
        </div>
        <span className={pill.role}>{session.user.email}</span>
      </div>

      <SellerProductForm
        submitLabel="Create product"
        submittingLabel="Creating..."
        onSubmit={createProduct}
        error={error}
      />
    </section>
  );
}
