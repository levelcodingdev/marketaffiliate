export const shell = {
  app: 'min-h-screen bg-gradient-to-b from-white to-slate-100 text-slate-900',
  topbar: 'sticky top-0 z-10 grid grid-cols-1 items-center gap-4 border-b border-slate-200/70 bg-white/90 px-4 py-3 backdrop-blur-md md:grid-cols-[auto_1fr_auto] md:gap-6 md:px-12',
  brand: 'inline-flex items-center gap-2.5 whitespace-nowrap text-base font-extrabold text-slate-950',
  brandMark: 'grid h-8 w-8 place-items-center rounded-lg bg-slate-950 text-white',
  nav: 'flex flex-wrap items-center gap-1.5',
  navLink: 'rounded-md px-3 py-2 text-sm font-bold text-slate-600 transition-colors hover:bg-teal-50 hover:text-teal-700',
  navLinkActive: 'bg-teal-50 text-teal-700',
  account: 'flex flex-wrap items-center justify-start gap-2.5 md:justify-end',
  page: 'mx-auto px-0 py-9 md:py-10',
};

export const layout = {
  stack: 'grid gap-6',
  pageHeading: 'flex flex-col items-start justify-between gap-4 md:flex-row md:items-end',
  title: 'm-0 max-w-3xl text-3xl font-extrabold leading-tight tracking-normal text-slate-950',
  subtitle: 'mt-3 max-w-2xl text-slate-600',
  eyebrow: 'mb-2 text-xs font-extrabold uppercase tracking-normal text-teal-700',
  muted: 'text-slate-500',
};

export const button = {
  base: 'inline-flex min-h-10 items-center justify-center rounded-md border border-transparent px-4 text-sm font-extrabold leading-none transition hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-60 max-sm:w-full',
  primary: 'bg-teal-700 text-white hover:bg-teal-800',
  secondary: 'border-teal-200 bg-teal-50 text-teal-700 hover:bg-teal-100',
  ghost: 'border-slate-300 bg-white text-slate-700 hover:bg-slate-50',
};

export function buttonClass(variant = 'primary', extra = '') {
  return `${button.base} ${button[variant]} ${extra}`.trim();
}

export const pill = {
  role: 'inline-flex min-h-7 items-center whitespace-nowrap rounded-full bg-emerald-100 px-3 text-xs font-extrabold capitalize text-emerald-800',
  roleSmall: 'inline-flex min-h-6 items-center whitespace-nowrap rounded-full bg-emerald-100 px-2 text-xs font-extrabold capitalize text-emerald-800',
  commission: 'inline-flex min-h-7 items-center whitespace-nowrap rounded-full bg-orange-100 px-3 text-xs font-extrabold capitalize text-orange-900',
};

export const feedback = {
  alert: 'rounded-lg border border-red-200 bg-red-50 px-3.5 py-3 font-bold text-red-800',
  loading: 'rounded-lg border border-dashed border-slate-300 bg-white p-5 font-extrabold text-slate-500',
  empty: 'grid justify-items-start gap-2.5 rounded-lg border border-slate-200 bg-white p-7 shadow-sm',
};

export const form = {
  root: 'grid gap-4',
  grid: 'grid grid-cols-1 gap-3.5 md:grid-cols-[1.3fr_0.8fr_0.8fr]',
  label: 'grid gap-2 text-sm font-extrabold text-slate-600',
  control: 'min-h-11 w-full rounded-md border border-slate-300 bg-white px-3 text-slate-950 outline-none transition focus:border-teal-700 focus:ring-4 focus:ring-teal-700/10',
  textarea: 'min-h-24 w-full resize-y rounded-md border border-slate-300 bg-white px-3 py-3 text-slate-950 outline-none transition focus:border-teal-700 focus:ring-4 focus:ring-teal-700/10',
  footer: 'mb-0 text-center',
};

export const panel = {
  base: 'rounded-lg border border-slate-200 bg-white shadow-sm',
};

export const table = {
  shell: 'overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm',
  heading: 'flex items-center justify-between gap-4 border-b border-slate-200 px-5 py-4',
  headingTitle: 'm-0 text-lg font-extrabold leading-snug text-slate-950',
  headingMeta: 'text-sm font-extrabold text-slate-500',
  scroll: 'overflow-x-auto',
  table: 'w-full min-w-[720px] border-collapse',
  th: 'border-b border-slate-200 bg-slate-50 px-4 py-3.5 text-left align-middle text-xs font-extrabold uppercase tracking-normal text-slate-500',
  td: 'border-b border-slate-200 px-4 py-3.5 text-left align-middle text-sm text-slate-800',
};

export const metric = {
  grid: 'grid grid-cols-1 gap-3.5 md:grid-cols-2 lg:grid-cols-4',
  card: 'rounded-lg border border-slate-200 bg-white p-4 shadow-sm',
  label: 'mb-2 block text-sm font-extrabold text-slate-500',
  value: 'block break-words text-3xl font-extrabold leading-none text-slate-950',
  valueCompact: 'block break-words text-lg font-extrabold leading-tight text-slate-950',
};
