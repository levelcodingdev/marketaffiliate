import { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../services/api.js';
import { money, percent } from '../services/format.js';
import { buttonClass, feedback, table } from '../ui/classes.js';

export default function SellerProductManager({ products, onChanged }) {
  const [deletingId, setDeletingId] = useState(null);
  const [error, setError] = useState('');

  const remove = async (product) => {
    if (!window.confirm(`Delete "${product.name}"?`)) return;

    setDeletingId(product.id);
    setError('');
    try {
      await api(`/products/${product.id}`, { method: 'DELETE' });
      await onChanged();
    } catch (err) {
      setError(err.message);
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <section className={table.shell}>
      <div className={table.heading}>
        <div>
          <h2 className={table.headingTitle}>Product catalog</h2>
          <p className="m-0 mt-1 text-sm text-slate-500">
            {products.length ? `${products.length} live ${products.length === 1 ? 'listing' : 'listings'}` : 'No listings yet'}
          </p>
        </div>
        <Link className={buttonClass('primary')} to="/seller/products/new">Create product</Link>
      </div>

      {error ? <div className="px-5 pt-5"><div className={feedback.alert}>{error}</div></div> : null}

      <div className={table.scroll}>
        <table className={table.table}>
          <thead>
            <tr>
              <th className={table.th}>Product</th>
              <th className={table.th}>Price</th>
              <th className={table.th}>Commission</th>
              <th className={table.th}>Sales</th>
              <th className={table.th}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {products.map((product) => (
              <tr className="hover:bg-slate-50" key={product.id}>
                <td className={table.td}>
                  <div className="font-bold text-slate-950">{product.name}</div>
                  <div className="max-w-md truncate text-xs text-slate-500">{product.description || 'No description'}</div>
                </td>
                <td className={table.td}>{money(product.price_cents)}</td>
                <td className={table.td}>{percent(product.commission_percent)}</td>
                <td className={table.td}>{product.sales || 0}</td>
                <td className={`${table.td} whitespace-nowrap`}>
                  <Link className={buttonClass('secondary', 'mr-2 min-h-8 px-3')} to={`/seller/products/${product.id}/edit`}>Edit</Link>
                  <button
                    className={buttonClass('ghost', 'min-h-8 px-3 text-red-700 hover:bg-red-50')}
                    type="button"
                    onClick={() => remove(product)}
                    disabled={deletingId === product.id}
                  >
                    {deletingId === product.id ? 'Deleting...' : 'Delete'}
                  </button>
                </td>
              </tr>
            ))}
            {!products.length ? (
              <tr>
                <td className={`${table.td} py-8 text-center`} colSpan="5">
                  <div className="grid justify-items-center gap-3">
                    <div>
                      <p className="m-0 font-extrabold text-slate-950">Start your catalog</p>
                      <p className="m-0 mt-1 text-slate-500">Create your first product so customers can find it in the marketplace.</p>
                    </div>
                    <Link className={buttonClass('primary')} to="/seller/products/new">Create product</Link>
                  </div>
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </section>
  );
}
