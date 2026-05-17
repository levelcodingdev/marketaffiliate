import { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { api } from '../../services/api.js';
import { money, percent } from '../../services/format.js';
import { captureReferral, getReferralCode } from '../../services/referrals.js';
import { buttonClass, feedback, layout, panel } from '../../ui/classes.js';

export default function ProductDetailPage() {
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const referralCode = searchParams.get('ref');
  const checkoutStatus = searchParams.get('checkout');
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [buying, setBuying] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    captureReferral(id, referralCode);
    if (referralCode) {
      api(`/track?ref=${encodeURIComponent(referralCode)}&product_id=${encodeURIComponent(id)}`).catch(() => {});
    }
  }, [id, referralCode]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError('');

    api(`/products/${id}`)
      .then((data) => {
        if (active) setProduct(data.product);
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
  }, [id]);

  const buy = async () => {
    setBuying(true);
    setError('');

    try {
      const data = await api('/checkout-session', {
        method: 'POST',
        body: {
          product_id: Number(id),
          referral_code: referralCode || getReferralCode(),
        },
      });
      window.location.assign(data.checkout_url);
    } catch (err) {
      setError(err.message);
      setBuying(false);
    }
  };

  const checkoutMessage = checkoutStatus === 'success'
    ? 'Payment complete. Your conversion will appear after the Stripe webhook is received.'
    : '';

  return (
    <section className="grid gap-5">
      {loading ? <div className={feedback.loading}>Loading product...</div> : null}
      {error ? <div className={feedback.alert}>{error}</div> : null}
      {checkoutStatus === 'cancelled' ? <div className="rounded-lg border border-amber-200 bg-amber-50 px-3.5 py-3 font-bold text-amber-800">Checkout cancelled. No payment was recorded.</div> : null}
      {checkoutMessage ? <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3.5 py-3 font-bold text-emerald-800">{checkoutMessage}</div> : null}

      {product ? (
        <>
          <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
            <div className="py-0 lg:py-6">
              <p className={layout.eyebrow}>Marketplace product</p>
              <h1 className="mb-4 text-4xl font-extrabold leading-tight text-slate-950">{product.name}</h1>
              <p className="max-w-3xl text-base text-slate-600">{product.description || 'No product description has been added yet.'}</p>
            </div>

            <aside className={`${panel.base} grid gap-3.5 p-5`}>
              <span className="text-sm font-extrabold text-slate-500">Price</span>
              <strong className="text-3xl font-extrabold text-slate-950">{money(product.price_cents)}</strong>
              <p className="mb-0 text-slate-600">{percent(product.commission_percent)} affiliate commission</p>
              <button className={buttonClass('primary', 'w-full')} type="button" onClick={buy} disabled={buying}>
                {buying ? 'Opening Stripe checkout...' : 'Buy with Stripe'}
              </button>
              <p className="m-0 text-xs font-bold text-slate-500">Secure checkout handled by Stripe.</p>
            </aside>
          </div>
        </>
      ) : null}
    </section>
  );
}
