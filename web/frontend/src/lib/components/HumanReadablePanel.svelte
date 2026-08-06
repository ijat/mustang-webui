<script lang="ts">
  import Icon from './Icon.svelte';
  import { formatDate, formatMoney } from '../format';
  import type { Invoice } from '../types';

  let { invoice }: { invoice: Invoice | null } = $props();
</script>

{#if !invoice}
  <div class="flex flex-col items-center px-8 py-16 text-center">
    <Icon name="file-x" class="mb-4 h-[34px] w-[34px] text-faint" />
    <p class="m-0 mb-2 text-lg font-medium text-ink">No embedded invoice data</p>
    <p class="m-0 max-w-[46ch] text-base text-muted">
      mustangproject didn't find ZUGFeRD/Factur-X XML embedded in this PDF, so there's no structured invoice
      to show. It may be a plain PDF/A document, or the XML attachment is missing entirely — check the raw
      XML tab and the validation checks for detail.
    </p>
  </div>
{:else}
  <div class="mb-[26px] grid grid-cols-1 gap-6 sm:grid-cols-2">
    <div>
      <h4 class="m-0 mb-2 text-xs font-medium uppercase tracking-[0.08em] text-muted">Seller</h4>
      <p class="m-0 mb-0.5 text-md font-semibold text-ink">{invoice.seller.name}</p>
      <p class="m-0 mb-0.5 text-base text-muted">
        {invoice.seller.street}, {invoice.seller.zip} {invoice.seller.city}, {invoice.seller.country}
      </p>
      <p class="m-0 text-base text-muted">VAT {invoice.seller.vatId}</p>
    </div>
    <div>
      <h4 class="m-0 mb-2 text-xs font-medium uppercase tracking-[0.08em] text-muted">Buyer</h4>
      <p class="m-0 mb-0.5 text-md font-semibold text-ink">{invoice.buyer.name}</p>
      <p class="m-0 mb-0.5 text-base text-muted">
        {invoice.buyer.street}, {invoice.buyer.zip} {invoice.buyer.city}, {invoice.buyer.country}
      </p>
      <p class="m-0 text-base text-muted">VAT {invoice.buyer.vatId}</p>
    </div>
  </div>

  <dl class="mb-[26px] grid grid-cols-2 gap-4 border border-hairline bg-surface-2 px-[18px] py-4 sm:grid-cols-3">
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Invoice no.</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{invoice.number}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Issue date</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{formatDate(invoice.issueDate)}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Due date</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{formatDate(invoice.dueDate)}</dd>
    </div>
    {#if invoice.deliveryDate}
      <div>
        <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Delivery date</dt>
        <dd class="m-0 text-md font-medium tabular-nums text-ink">{formatDate(invoice.deliveryDate)}</dd>
      </div>
    {/if}
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Currency</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{invoice.currency}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Payment terms</dt>
      <dd class="m-0 text-md font-medium text-ink">{invoice.paymentTerms ?? '—'}</dd>
    </div>
    {#if invoice.buyerReference}
      <div>
        <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Buyer reference</dt>
        <dd class="m-0 text-md font-medium text-ink">{invoice.buyerReference}</dd>
      </div>
    {/if}
  </dl>

  {#if invoice.paymentMeans || invoice.paymentReference}
    <dl class="mb-[26px] grid grid-cols-2 gap-4 border border-hairline bg-surface-2 px-[18px] py-4 sm:grid-cols-3">
      {#if invoice.paymentMeans?.iban}
        <div>
          <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">IBAN</dt>
          <dd class="m-0 text-md font-medium tabular-nums text-ink">{invoice.paymentMeans.iban}</dd>
        </div>
      {/if}
      {#if invoice.paymentMeans?.bic}
        <div>
          <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">BIC</dt>
          <dd class="m-0 text-md font-medium tabular-nums text-ink">{invoice.paymentMeans.bic}</dd>
        </div>
      {/if}
      {#if invoice.paymentMeans?.accountName}
        <div>
          <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Account holder</dt>
          <dd class="m-0 text-md font-medium text-ink">{invoice.paymentMeans.accountName}</dd>
        </div>
      {/if}
      {#if invoice.paymentReference}
        <div>
          <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Payment reference</dt>
          <dd class="m-0 text-md font-medium text-ink">{invoice.paymentReference}</dd>
        </div>
      {/if}
    </dl>
  {/if}

  <div class="overflow-x-auto border border-hairline">
    <table class="w-full min-w-[560px] border-collapse text-base">
      <thead>
        <tr>
          <th class="border-b border-hairline bg-surface-2 px-3.5 py-2.5 text-left text-xs font-medium uppercase tracking-[0.05em] text-muted">Description</th>
          <th class="border-b border-hairline bg-surface-2 px-3.5 py-2.5 text-right text-xs font-medium uppercase tracking-[0.05em] text-muted">Qty</th>
          <th class="border-b border-hairline bg-surface-2 px-3.5 py-2.5 text-right text-xs font-medium uppercase tracking-[0.05em] text-muted">Unit price</th>
          <th class="border-b border-hairline bg-surface-2 px-3.5 py-2.5 text-right text-xs font-medium uppercase tracking-[0.05em] text-muted">VAT</th>
          <th class="border-b border-hairline bg-surface-2 px-3.5 py-2.5 text-right text-xs font-medium uppercase tracking-[0.05em] text-muted">Total</th>
        </tr>
      </thead>
      <tbody>
        {#each invoice.lineItems as item, i (i)}
          <tr>
            <td class="border-b border-hairline px-3.5 py-2.5 align-top">{item.description}</td>
            <td class="border-b border-hairline px-3.5 py-2.5 text-right align-top tabular-nums">{item.quantity}</td>
            <td class="border-b border-hairline px-3.5 py-2.5 text-right align-top tabular-nums">{formatMoney(item.unitPrice, invoice.currency)}</td>
            <td class="border-b border-hairline px-3.5 py-2.5 text-right align-top tabular-nums">{item.vatPercent}%</td>
            <td class="border-b border-hairline px-3.5 py-2.5 text-right align-top tabular-nums">{formatMoney(item.lineTotal, invoice.currency)}</td>
          </tr>
        {/each}
      </tbody>
      <tfoot>
        {#each invoice.allowances as allowance, i (i)}
          <tr>
            <td colspan="4" class="border-t border-hairline px-3.5 py-2.5 font-sans tabular-nums text-muted">
              Allowance{allowance.reason ? `: ${allowance.reason}` : ''}{allowance.percent != null ? ` (${allowance.percent}%)` : ''}
            </td>
            <td class="border-t border-hairline px-3.5 py-2.5 text-right font-sans tabular-nums text-muted">
              {allowance.amount != null ? `−${formatMoney(allowance.amount, invoice.currency)}` : '—'}
            </td>
          </tr>
        {/each}
        {#each invoice.charges as charge, i (i)}
          <tr>
            <td colspan="4" class="border-t border-hairline px-3.5 py-2.5 font-sans tabular-nums text-muted">
              Charge{charge.reason ? `: ${charge.reason}` : ''}{charge.percent != null ? ` (${charge.percent}%)` : ''}
            </td>
            <td class="border-t border-hairline px-3.5 py-2.5 text-right font-sans tabular-nums text-muted">
              {charge.amount != null ? `+${formatMoney(charge.amount, invoice.currency)}` : '—'}
            </td>
          </tr>
        {/each}
        <tr>
          <td colspan="4" class="border-t border-hairline px-3.5 py-2.5 font-sans tabular-nums">Net total</td>
          <td class="border-t border-hairline px-3.5 py-2.5 text-right font-sans tabular-nums">{formatMoney(invoice.totals.netTotal, invoice.currency)}</td>
        </tr>
        <tr>
          <td colspan="4" class="border-t border-hairline px-3.5 py-2.5 font-sans tabular-nums">VAT total</td>
          <td class="border-t border-hairline px-3.5 py-2.5 text-right font-sans tabular-nums">{formatMoney(invoice.totals.vatTotal, invoice.currency)}</td>
        </tr>
        <tr>
          <td colspan="4" class="border-t border-ink px-3.5 py-2.5 font-sans text-md font-semibold tabular-nums">Gross total</td>
          <td class="border-t border-ink px-3.5 py-2.5 text-right font-sans text-md font-semibold tabular-nums">{formatMoney(invoice.totals.grossTotal, invoice.currency)}</td>
        </tr>
      </tfoot>
    </table>
  </div>

  {#if invoice.notes.length > 0}
    <div class="mt-[26px]">
      <h4 class="m-0 mb-2 text-xs font-medium uppercase tracking-[0.08em] text-muted">Notes</h4>
      <ul class="m-0 flex flex-col gap-1.5 p-0">
        {#each invoice.notes as note, i (i)}
          <li class="list-none text-base text-muted">{note}</li>
        {/each}
      </ul>
    </div>
  {/if}
{/if}
