import { execSync } from 'child_process';
import { existsSync, mkdirSync, readdirSync } from 'fs';
import { join } from 'path';
import * as os from 'os';
import * as fs from 'fs';
import * as https from 'https';

/**
 * Script to generate SVG files from D2 diagram files
 * This allows comparison between diagram-dsl and D2 outputs
 */

const D2_VERSION = 'v0.6.7';
const D2_DIR = join(__dirname, '../../examples/d2');
const D2_OUTPUT_DIR = join(__dirname, '../../examples/d2-output');

/**
 * Downloads D2 from GitHub releases if not already installed
 */
async function ensureD2Installed(): Promise<string> {
  // Check if d2 is already in PATH
  try {
    execSync('which d2', { stdio: 'ignore' });
    console.log('✓ D2 is already installed');
    return 'd2';
  } catch {
    // D2 not in PATH, need to download
  }

  const platform = os.platform();
  const arch = os.arch();
  
  let d2Binary: string;
  
  if (platform === 'darwin') {
    d2Binary = arch === 'arm64' ? 'd2-v0.6.7-macos-arm64' : 'd2-v0.6.7-macos-amd64';
  } else if (platform === 'linux') {
    d2Binary = 'd2-v0.6.7-linux-amd64';
  } else if (platform === 'win32') {
    d2Binary = 'd2-v0.6.7-windows-amd64';
  } else {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  const downloadUrl = `https://github.com/terrastruct/d2/releases/download/${D2_VERSION}/${d2Binary}.tar.gz`;
  const tmpDir = fs.mkdtempSync(join(os.tmpdir(), 'd2-'));
  const tarFile = join(tmpDir, 'd2.tar.gz');
  const d2Path = join(tmpDir, 'd2-v0.6.7', 'bin', 'd2');

  console.log(`Downloading D2 ${D2_VERSION} for ${platform}/${arch}...`);
  
  return new Promise<string>((resolve, reject) => {
    https.get(downloadUrl, { followAllRedirects: true } as any, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        https.get(response.headers.location!, (redirectResponse) => {
          const file = fs.createWriteStream(tarFile);
          redirectResponse.pipe(file);
          file.on('finish', () => {
            file.close();
            console.log('✓ Downloaded D2');
            
            // Extract tar.gz
            console.log('Extracting D2...');
            try {
              execSync(`tar -xzf ${tarFile} -C ${tmpDir}`, { stdio: 'ignore' });
              console.log('✓ Extracted D2');
              
              // Make executable
              fs.chmodSync(d2Path, '755');
              
              resolve(d2Path);
            } catch (error) {
              reject(new Error(`Failed to extract D2: ${error}`));
            }
          });
        }).on('error', reject);
      } else {
        const file = fs.createWriteStream(tarFile);
        response.pipe(file);
        file.on('finish', () => {
          file.close();
          console.log('✓ Downloaded D2');
          
          // Extract tar.gz
          console.log('Extracting D2...');
          try {
            execSync(`tar -xzf ${tarFile} -C ${tmpDir}`, { stdio: 'ignore' });
            console.log('✓ Extracted D2');
            
            // Make executable
            fs.chmodSync(d2Path, '755');
            
            resolve(d2Path);
          } catch (error) {
            reject(new Error(`Failed to extract D2: ${error}`));
          }
        });
      }
    }).on('error', reject);
  });
}

/**
 * Generates SVG files from D2 diagram files
 */
async function generateD2Examples(d2Command: string): Promise<void> {
  console.log('\nGenerating D2 SVG examples...\n');

  // Create output directory if it doesn't exist
  if (!existsSync(D2_OUTPUT_DIR)) {
    mkdirSync(D2_OUTPUT_DIR, { recursive: true });
  }

  // Get all .d2 files
  const d2Files = readdirSync(D2_DIR)
    .filter(file => file.endsWith('.d2'))
    .sort();

  let successCount = 0;
  let failCount = 0;

  for (const d2File of d2Files) {
    const inputPath = join(D2_DIR, d2File);
    const filename = d2File.replace('.d2', '');
    const outputPath = join(D2_OUTPUT_DIR, `${filename}.svg`);

    try {
      console.log(`  Generating ${filename}.svg from D2...`);
      
      // Run D2 with default theme (no --theme flag)
      execSync(`"${d2Command}" "${inputPath}" "${outputPath}"`, {
        stdio: 'pipe',
      });

      if (existsSync(outputPath)) {
        console.log(`  ✓ Generated ${filename}.svg`);
        successCount++;
      } else {
        console.log(`  ✗ Failed to generate ${filename}.svg`);
        failCount++;
      }
    } catch (error) {
      console.error(`  ✗ Error generating ${filename}.svg:`, error);
      failCount++;
    }
  }

  console.log('\n' + '='.repeat(50));
  console.log(`D2 examples generated successfully!`);
  console.log(`  Success: ${successCount}`);
  console.log(`  Failed: ${failCount}`);
  console.log(`  Output directory: ${D2_OUTPUT_DIR}`);
  console.log('='.repeat(50) + '\n');
}

/**
 * Main function
 */
async function main() {
  try {
    const d2Command = await ensureD2Installed();
    await generateD2Examples(d2Command);
  } catch (error) {
    console.error('Error:', error);
    process.exit(1);
  }
}

main();
