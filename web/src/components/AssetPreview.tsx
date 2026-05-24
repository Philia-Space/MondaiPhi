import React from 'react';
import type { Asset } from '../api';

export function AssetPreview({ asset }: { asset: Asset }) {
  const url = asset.SourceURL || `/assets/${asset.ID}`;

  if (asset.Type === 'audio') {
    return <audio controls src={url} style={{ width: '220px', height: '32px' }} />;
  }

  if (asset.Type === 'image') {
    return (
      <a href={url} target="_blank" rel="noreferrer">
        <img src={url} alt={asset.ID} style={{ width: '96px', height: '64px', objectFit: 'cover', borderRadius: '6px', border: '1px solid #334155' }} />
      </a>
    );
  }

  return <a href={url} target="_blank" rel="noreferrer" style={{ color: '#60a5fa' }}>Open</a>;
}

export function AssetListPreview({ assets }: { assets: Asset[] }) {
  if (!assets.length) {
    return <span style={{ color: '#64748b' }}>-</span>;
  }

  return (
    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
      {assets.slice(0, 2).map((asset) => <AssetPreview key={asset.ID} asset={asset} />)}
      {assets.length > 2 && <span style={{ color: '#94a3b8', fontSize: '0.75rem' }}>+{assets.length - 2}</span>}
    </div>
  );
}
