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
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  const [buying, setBuying] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

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
    setSuccess('');

    try {
      const data = await api('/purchase', {
        method: 'POST',
        body: {
          product_id: Number(id),
          referral_code: referralCode || getReferralCode(),
        },
      });
      const reference = data.conversion?.payment_reference;
      setSuccess(reference ? `Test purchase recorded. Reference: ${reference}` : 'Test purchase recorded.');
    } catch (err) {
      setError(err.message);
    } finally {
      setBuying(false);
    }
  };

  const activeReferral = referralCode || getReferralCode();

  return (
    <section className="grid gap-5">
      {loading ? <div className={feedback.loading}>Loading product...</div> : null}
      {error ? <div className={feedback.alert}>{error}</div> : null}
      {success ? <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3.5 py-3 font-bold text-emerald-800">{success}</div> : null}

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
                {buying ? 'Recording purchase...' : 'Buy now (test)'}
              </button>
              <p className="m-0 text-xs font-bold text-slate-500">Instant test payment. No card required.</p>
            </aside>
          </div>

          <div className="flex flex-col items-start justify-between gap-3.5 rounded-lg border border-blue-200 bg-blue-50 px-4 py-3.5 text-blue-900 sm:flex-row sm:items-center">
            <span className="font-extrabold">Referral attribution</span>
            <strong className="break-words">{activeReferral || 'No active referral'}</strong>
          </div>
        </>
      ) : null}
    </section>
  );
}
