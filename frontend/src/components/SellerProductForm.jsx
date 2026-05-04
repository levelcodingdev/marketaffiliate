import { useState } from 'react';
import { Link } from 'react-router-dom';
import { buttonClass, feedback, form as formClasses, panel } from '../ui/classes.js';

const defaultValues = {
  name: '',
  description: '',
  price: '',
  commission_percent: 30,
};

export default function SellerProductForm({
  initialValues = defaultValues,
  submitLabel,
  submittingLabel,
  onSubmit,
  error,
}) {
  const [formState, setFormState] = useState(initialValues);
  const [saving, setSaving] = useState(false);

  const update = (event) => {
    setFormState((current) => ({ ...current, [event.target.name]: event.target.value }));
  };

  const submit = async (event) => {
    event.preventDefault();
    setSaving(true);

    try {
      await onSubmit({
        name: formState.name,
        description: formState.description,
        price: Number(formState.price),
        commission_percent: Number(formState.commission_percent),
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <form className={`${panel.base} ${formClasses.root} p-5`} onSubmit={submit}>
      {error ? <div className={feedback.alert}>{error}</div> : null}

      <div className={formClasses.grid}>
        <label className={formClasses.label}>
          Name
          <input className={formClasses.control} name="name" value={formState.name} onChange={update} />
        </label>
        <label className={formClasses.label}>
          Price
          <input className={formClasses.control} name="price" type="number" min="1" step="0.01" value={formState.price} onChange={update} />
        </label>
        <label className={formClasses.label}>
          Commission
          <input className={formClasses.control} name="commission_percent" type="number" min="0" max="100" value={formState.commission_percent} onChange={update} />
        </label>
      </div>

      <label className={formClasses.label}>
        Description
        <textarea className={formClasses.textarea} name="description" rows="5" value={formState.description} onChange={update} />
      </label>

      <div className="flex flex-col gap-3 sm:flex-row">
        <button className={buttonClass('primary')} type="submit" disabled={saving}>
          {saving ? submittingLabel : submitLabel}
        </button>
        <Link className={buttonClass('ghost')} to="/seller/products">Cancel</Link>
      </div>
    </form>
  );
}
