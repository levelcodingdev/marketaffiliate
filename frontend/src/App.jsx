import { Link, Navigate, NavLink, Route, Routes } from 'react-router-dom';
import { useEffect, useState } from 'react';
import DashboardPage from './pages/dashboard/DashboardPage.jsx';
import LoginPage from './pages/login/LoginPage.jsx';
import ProductDetailPage from './pages/product/ProductDetailPage.jsx';
import ProductsPage from './pages/products/ProductsPage.jsx';
import RegisterPage from './pages/register/RegisterPage.jsx';
import SellerProductCreatePage from './pages/seller/SellerProductCreatePage.jsx';
import SellerProductEditPage from './pages/seller/SellerProductEditPage.jsx';
import SellerProductsPage from './pages/seller/SellerProductsPage.jsx';
import { clearSession, getSession } from './services/session.js';
import { buttonClass, pill, shell } from './ui/classes.js';

export default function App() {
  const [session, setSession] = useState(getSession);

  useEffect(() => {
    const syncSession = () => setSession(getSession());
    window.addEventListener('storage', syncSession);
    window.addEventListener('session-change', syncSession);
    return () => {
      window.removeEventListener('storage', syncSession);
      window.removeEventListener('session-change', syncSession);
    };
  }, []);

  const logout = () => {
    clearSession();
  };

  return (
    <div className={shell.app}>
      <header className={shell.topbar}>
        <Link to="/products" className={shell.brand}>
          <span className={shell.brandMark}>A</span>
          <span>AffiliateTrack</span>
        </Link>

        <nav className={shell.nav}>
          <NavLink
            to="/products"
            className={({ isActive }) => `${shell.navLink} ${isActive ? shell.navLinkActive : ''}`}
          >
            Marketplace
          </NavLink>
          {session ? (
            <NavLink
              to="/dashboard"
              className={({ isActive }) => `${shell.navLink} ${isActive ? shell.navLinkActive : ''}`}
            >
              Dashboard
            </NavLink>
          ) : null}
          {session?.user?.role === 'seller' ? (
            <NavLink
              to="/seller/products"
              className={({ isActive }) => `${shell.navLink} ${isActive ? shell.navLinkActive : ''}`}
            >
              My Products
            </NavLink>
          ) : null}
        </nav>

        <div className={shell.account}>
          {session ? (
            <>
              <span className={pill.role}>{session.user.role}</span>
              <button className={buttonClass('secondary')} type="button" onClick={logout}>Logout</button>
            </>
          ) : (
            <>
              <Link className={buttonClass('ghost')} to="/login">Login</Link>
              <Link className={buttonClass('primary')} to="/register">Register</Link>
            </>
          )}
        </div>
      </header>

      <main className={shell.page} style={{ width: 'min(1180px, calc(100% - 32px))' }}>
        <Routes>
          <Route path="/" element={<Navigate to="/products" replace />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/products" element={<ProductsPage session={session} />} />
          <Route path="/product/:id" element={<ProductDetailPage session={session} />} />
          <Route path="/seller/products" element={<SellerProductsPage session={session} />} />
          <Route path="/seller/products/new" element={<SellerProductCreatePage session={session} />} />
          <Route path="/seller/products/:id/edit" element={<SellerProductEditPage session={session} />} />
          <Route path="/dashboard" element={<DashboardPage session={session} />} />
        </Routes>
      </main>
    </div>
  );
}
