import { createMDX } from 'fumadocs-mdx/next';
import type { NextConfig } from 'next';

const isProd = process.env.NODE_ENV === 'production';

const config: NextConfig = {
  reactStrictMode: true,
  output: 'export',
  basePath: isProd ? '/scfleet' : '',
  assetPrefix: isProd ? '/scfleet/' : '',
};

export default createMDX()(config);
