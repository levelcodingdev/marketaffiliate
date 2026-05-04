import { Link } from 'react-router-dom';
import { useCallback, useEffect, useState } from 'react';
import { api } from '../../services/api.js';
import { money, percent } from '../../services/format.js';
import { buttonClass, feedback, form as formClasses, layout, panel, pill } from '../../ui/classes.js';

export default function ProductsPage({ session }) {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');

  const referralCode = session?.user?.role === 'affiliate' ? session.user.referral_code : '';
  const isAffiliate = session?.user?.role === 'affiliate';
  const filteredProducts = products.filter((product) => {
    const query = search.trim().toLowerCase();
    if (!query) return true;
    return `${product.name} ${product.description}`.toLowerCase().includes(query);
  });

  const loadProducts = useCallback(async () => {
    setLoading(true);
    setError('');

    try {
      const data = await api('/products');
      setProducts(data.products || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProducts();
  }, [loadProducts]);

  return (
    <section className={layout.stack}>
      <div className={layout.pageHeading}>
        <div>
          <p className={layout.eyebrow}>Public marketplace</p>
          <h1 className={layout.title}>Discover products from independent sellers</h1>
          <p className={layout.subtitle}>Browse digital products, templates, courses, and tools from sellers on AffiliateTrack.</p>
        </div>
      </div>

      <div className={`${panel.base} flex items-center justify-between p-3.5`}>
        <label className="grid w-full max-w-xl gap-2 text-sm font-extrabold text-slate-600">
          <span>Search products</span>
          <input
            className={formClasses.control}
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search by product or keyword"
          />
        </label>
      </div>

      {error ? <div className={feedback.alert}>{error}</div> : null}
      {loading ? <div className={feedback.loading}>Loading marketplace...</div> : null}

      {!loading && !products.length ? (
        <div className={feedback.empty}>
          <h2 className="m-0 text-lg font-extrabold text-slate-950">No products yet</h2>
          <p className="m-0">Products published by sellers will appear here.</p>
        </div>
      ) : null}

      {!loading && products.length > 0 && !filteredProducts.length ? (
        <div className={feedback.empty}>
          <h2 className="m-0 text-lg font-extrabold text-slate-950">No matching products</h2>
          <p className="m-0">Try another search term.</p>
        </div>
      ) : null}

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {filteredProducts.map((product) => {
          const link = referralCode
            ? `/product/${product.id}?ref=${encodeURIComponent(referralCode)}`
            : `/product/${product.id}`;

          return (
            <article className="flex min-h-[360px] flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-teal-200 hover:shadow-lg" key={product.id}>
              <div className="grid min-h-32 place-items-center border-b border-slate-200 bg-gradient-to-br from-teal-100 to-slate-100">
                <span className="grid h-14 w-14 place-items-center rounded-lg border border-teal-700/20 bg-white/75 text-lg font-black text-teal-700 shadow-md">{product.name.slice(0, 2).toUpperCase()}</span>
              </div>

              <div className="grid flex-1 content-start gap-4 px-4 pt-4">
                <div className="flex min-h-8 flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
                  <span className="break-words text-xs font-extrabold uppercase text-teal-700">Marketplace pick</span>
                  {isAffiliate ? <span className={pill.commission}>{percent(product.commission_percent)} commission</span> : null}
                </div>

                <div>
                  <h2 className="mb-2 text-xl font-extrabold leading-snug text-slate-950">{product.name}</h2>
                  <p className="line-clamp-3-manual mb-0 min-h-[72px] overflow-hidden text-slate-600">{product.description || 'No description provided.'}</p>
                </div>
              </div>

              <div className="mt-auto flex flex-col items-start justify-between gap-3 border-t border-slate-200 bg-slate-50 px-4 py-4 sm:flex-row sm:items-center">
                <div className="grid gap-1">
                  <span className="text-xs font-extrabold text-slate-500">One-time purchase</span>
                  <strong className="text-2xl font-extrabold leading-none text-slate-950">{money(product.price_cents)}</strong>
                </div>
                <Link className={buttonClass('primary')} to={link}>{isAffiliate ? 'Get link' : 'View product'}</Link>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
