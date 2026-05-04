import { Link } from 'react-router-dom';
import { useCallback, useEffect, useState } from 'react';
import { api } from '../../services/api.js';
import { money, percent } from '../../services/format.js';
import { buttonClass, feedback, layout, metric, panel, pill, table } from '../../ui/classes.js';

export default function DashboardPage({ session }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(Boolean(session));
  const [error, setError] = useState('');

  const loadDashboard = useCallback(async () => {
    if (!session) return;

    setLoading(true);
    setError('');

    try {
      const payload = await api(`/dashboard/${session.user.role}`);
      setData(payload);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [session]);

  useEffect(() => {
    loadDashboard();
  }, [loadDashboard]);

  if (!session) {
    return (
      <section className={feedback.empty}>
        <h1 className="m-0 text-3xl font-extrabold text-slate-950">Sign in to view your dashboard</h1>
        <p className="m-0 text-slate-600">Your dashboard is personalized by role.</p>
        <Link className={buttonClass('primary')} to="/login">Login</Link>
      </section>
    );
  }

  return (
    <section className={layout.stack}>
      <div className={layout.pageHeading}>
        <div>
          <p className={layout.eyebrow}>{session.user.role} dashboard</p>
          <h1 className={layout.title}>{dashboardTitle(session.user.role)}</h1>
        </div>
        <span className={pill.role}>{session.user.email}</span>
      </div>

      {loading ? <div className={feedback.loading}>Loading dashboard...</div> : null}
      {error ? <div className={feedback.alert}>{error}</div> : null}

      {data && session.user.role === 'affiliate' ? <AffiliateDashboard data={data} /> : null}
      {data && session.user.role === 'seller' ? <SellerDashboard data={data} /> : null}
      {data && session.user.role === 'admin' ? <AdminDashboard data={data} /> : null}
    </section>
  );
}

function AffiliateDashboard({ data }) {
  return (
    <>
      <div className={metric.grid}>
        <Metric label="Clicks" value={data.clicks} />
        <Metric label="Conversions" value={data.conversions} />
        <Metric label="Earnings" value={money(data.earnings_cents)} />
        <Metric label="Referral code" value={data.referral_code} compact />
      </div>

      <section className={table.shell}>
        <div className={table.heading}>
          <h2 className={table.headingTitle}>Referral links</h2>
          <span className={table.headingMeta}>{data.products?.length || 0} products</span>
        </div>
        <div className={table.scroll}>
          <table className={table.table}>
            <thead>
              <tr>
                <th className={table.th}>Product</th>
                <th className={table.th}>Seller</th>
                <th className={table.th}>Price</th>
                <th className={table.th}>Commission</th>
                <th className={table.th}>Link</th>
              </tr>
            </thead>
            <tbody>
              {(data.products || []).map((product) => (
                <tr className="hover:bg-slate-50" key={product.id}>
                  <td className={table.td}>{product.name}</td>
                  <td className={table.td}>{product.seller_email}</td>
                  <td className={table.td}>{money(product.price_cents)}</td>
                  <td className={table.td}>{percent(product.commission_percent)}</td>
                  <td className={table.td}><a className="font-bold text-teal-700 no-underline" href={product.referral_url}>Open</a></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}

function SellerDashboard({ data }) {
  return (
    <>
      <div className={metric.grid}>
        <Metric label="Products" value={data.products_count} />
        <Metric label="Sales" value={data.sales} />
        <Metric label="Gross revenue" value={money(data.gross_revenue_cents)} />
        <Metric label="Net revenue" value={money(data.net_revenue_cents)} />
      </div>

      <section className={`${panel.base} flex flex-col justify-between gap-4 p-5 md:flex-row md:items-center`}>
        <div>
          <h2 className="mb-2 text-lg font-extrabold leading-snug text-slate-950">Product catalog</h2>
          <p className="m-0 text-slate-600">Create, edit, and delete products from the seller products page.</p>
        </div>
        <Link className={buttonClass('primary')} to="/seller/products">Manage products</Link>
      </section>

      <section className={table.shell}>
        <div className={table.heading}>
          <h2 className={table.headingTitle}>Product performance</h2>
          <span className={table.headingMeta}>{money(data.commission_paid_cents)} commissions paid</span>
        </div>
        <div className={table.scroll}>
          <table className={table.table}>
            <thead>
              <tr>
                <th className={table.th}>Product</th>
                <th className={table.th}>Price</th>
                <th className={table.th}>Commission</th>
                <th className={table.th}>Sales</th>
                <th className={table.th}>Gross</th>
                <th className={table.th}>Net</th>
              </tr>
            </thead>
            <tbody>
              {(data.products || []).map((product) => (
                <tr className="hover:bg-slate-50" key={product.id}>
                  <td className={table.td}>{product.name}</td>
                  <td className={table.td}>{money(product.price_cents)}</td>
                  <td className={table.td}>{percent(product.commission_percent)}</td>
                  <td className={table.td}>{product.sales}</td>
                  <td className={table.td}>{money(product.gross_revenue_cents)}</td>
                  <td className={table.td}>{money(product.net_revenue_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className={table.shell}>
        <div className={table.heading}>
          <h2 className={table.headingTitle}>Affiliate performance</h2>
          <span className={table.headingMeta}>{data.affiliates?.length || 0} active affiliates</span>
        </div>
        <div className={table.scroll}>
          <table className={table.table}>
            <thead>
              <tr>
                <th className={table.th}>Affiliate</th>
                <th className={table.th}>Conversions</th>
                <th className={table.th}>Revenue</th>
                <th className={table.th}>Commission</th>
              </tr>
            </thead>
            <tbody>
              {(data.affiliates || []).map((affiliate) => (
                <tr className="hover:bg-slate-50" key={affiliate.affiliate_id}>
                  <td className={table.td}>{affiliate.affiliate_email}</td>
                  <td className={table.td}>{affiliate.conversions}</td>
                  <td className={table.td}>{money(affiliate.revenue_cents)}</td>
                  <td className={table.td}>{money(affiliate.commission_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}

function AdminDashboard({ data }) {
  return (
    <>
      <div className={metric.grid}>
        <Metric label="Users" value={data.users_count} />
        <Metric label="Products" value={data.products_count} />
        <Metric label="Revenue" value={money(data.gross_revenue_cents)} />
        <Metric label="Commissions" value={money(data.commission_cents)} />
      </div>

      <section className={table.shell}>
        <div className={table.heading}>
          <h2 className={table.headingTitle}>Users</h2>
          <span className={table.headingMeta}>{data.users?.length || 0} accounts</span>
        </div>
        <div className={table.scroll}>
          <table className={table.table}>
            <thead>
              <tr>
                <th className={table.th}>Email</th>
                <th className={table.th}>Role</th>
                <th className={table.th}>Referral code</th>
                <th className={table.th}>Created</th>
              </tr>
            </thead>
            <tbody>
              {(data.users || []).map((user) => (
                <tr className="hover:bg-slate-50" key={user.id}>
                  <td className={table.td}>{user.email}</td>
                  <td className={table.td}><span className={pill.roleSmall}>{user.role}</span></td>
                  <td className={table.td}>{user.referral_code}</td>
                  <td className={table.td}>{new Date(user.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className={table.shell}>
        <div className={table.heading}>
          <h2 className={table.headingTitle}>Transactions</h2>
          <span className={table.headingMeta}>{data.transactions?.length || 0} conversions</span>
        </div>
        <div className={table.scroll}>
          <table className={table.table}>
            <thead>
              <tr>
                <th className={table.th}>Product</th>
                <th className={table.th}>Seller</th>
                <th className={table.th}>Affiliate</th>
                <th className={table.th}>Amount</th>
                <th className={table.th}>Commission</th>
              </tr>
            </thead>
            <tbody>
              {(data.transactions || []).map((transaction) => (
                <tr className="hover:bg-slate-50" key={transaction.id}>
                  <td className={table.td}>{transaction.product_name}</td>
                  <td className={table.td}>{transaction.seller_email}</td>
                  <td className={table.td}>{transaction.affiliate_email || 'Direct'}</td>
                  <td className={table.td}>{money(transaction.amount_cents)}</td>
                  <td className={table.td}>{money(transaction.commission_amount_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </>
  );
}

function Metric({ label, value, compact = false }) {
  return (
    <article className={metric.card}>
      <span className={metric.label}>{label}</span>
      <strong className={compact ? metric.valueCompact : metric.value}>{value}</strong>
    </article>
  );
}

function dashboardTitle(role) {
  if (role === 'seller') return 'Sales and product performance';
  if (role === 'admin') return 'Platform overview';
  return 'Clicks, conversions, and earnings';
}
