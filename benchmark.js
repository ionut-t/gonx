const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const { format } = require('date-fns');

// Console colors using ANSI escape codes
const colors = {
    reset: '\x1b[0m',
    bright: '\x1b[1m',
    dim: '\x1b[2m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    magenta: '\x1b[35m',
    cyan: '\x1b[36m',
    white: '\x1b[37m',
};

class NxBenchmark {
    constructor() {
        this.metricsHistory = [];
        this.metricsPath = 'benchmark-metrics.json';
        this.loadMetricsHistory();
    }

    loadMetricsHistory() {
        try {
            if (fs.existsSync(this.metricsPath)) {
                this.metricsHistory = JSON.parse(fs.readFileSync(this.metricsPath, 'utf8'));
            }
        } catch (error) {
            console.error(`${colors.red}Error loading metrics history:${colors.reset}`, error);
        }
    }

    saveMetricsHistory() {
        try {
            fs.writeFileSync(this.metricsPath, JSON.stringify(this.metricsHistory, null, 2));
        } catch (error) {
            console.error(`${colors.red}Error saving metrics history:${colors.reset}`, error);
        }
    }

    getAllApps() {
        try {
            const output = execSync('nx show projects --json', { encoding: 'utf8' });
            const projects = JSON.parse(output);

            return projects.filter(project => {
                try {
                    const projectConfig = JSON.parse(execSync(`nx show project ${project} --json`, { encoding: 'utf8' }));

                    // Exclude e2e projects
                    if (
                        project.endsWith('-e2e') ||
                        project.includes('.e2e') ||
                        projectConfig.root.includes('/e2e') ||
                        projectConfig.tags?.includes('e2e')
                    ) {
                        return false;
                    }

                    // Include only actual applications (not libraries or e2e projects)
                    return (
                        projectConfig.projectType === 'application' &&
                        !projectConfig.root.includes('/libs/') &&
                        projectConfig.targets?.build
                    );
                } catch (error) {
                    console.error(`${colors.dim}Skipping ${project}: ${error.message}${colors.reset}`);
                    return false;
                }
            });
        } catch (error) {
            console.error(`${colors.red}Error getting apps list:${colors.reset}`, error);
            return [];
        }
    }

    calculateBundleSize(appName) {
        const stats = {
            initial: {
                main: 0,
                runtime: 0,
                polyfills: 0,
                total: 0,
            },
            lazy: 0,
            assets: 0,
            total: 0,
            totalWithAssets: 0,
        };

        try {
            const appPath = path.join(process.cwd(), 'dist', 'apps', appName, 'browser');
            const assetsPath = path.join(appPath, 'assets');

            console.log(`\n${colors.cyan}📁 Looking for files in:${colors.reset} ${colors.yellow}${appPath}${colors.reset}`);

            if (!fs.existsSync(appPath)) {
                console.error(`${colors.red}❌ Build output directory not found:${colors.reset} ${appPath}`);
                return stats;
            }

            // Calculate assets size
            if (fs.existsSync(assetsPath)) {
                const calculateDirSize = dirPath => {
                    let size = 0;
                    const files = fs.readdirSync(dirPath);

                    for (const file of files) {
                        const filePath = path.join(dirPath, file);
                        const stat = fs.statSync(filePath);

                        if (stat.isDirectory()) {
                            size += calculateDirSize(filePath);
                        } else {
                            size += stat.size;
                        }
                    }

                    return size;
                };

                stats.assets = calculateDirSize(assetsPath);
                console.log(`\n${colors.cyan}📂 Assets size:${colors.reset} ${this.formatSize(stats.assets)}`);
            }

            console.log(`\n${colors.cyan}📊 Scanning bundle files:${colors.reset}`);
            console.log(`${colors.cyan}------------------------------${colors.reset}`);

            const files = fs.readdirSync(appPath);

            files.forEach(file => {
                // Skip source maps and non-JS files
                const isSourceMap = file.endsWith('.map');
                if (!file.endsWith('.js') || isSourceMap) {
                    const fileDescription = isSourceMap ? 'source map' : 'not a JavaScript file';

                    console.log(`${colors.dim}→ Skipping: ${file} (${fileDescription})${colors.reset}`);
                    return;
                }

                const fullPath = path.join(appPath, file);
                const size = fs.statSync(fullPath).size;

                console.log(`${colors.green}File size: ${(size / 1024).toFixed(2)} KB${colors.reset}`);

                if (file.startsWith('main-')) {
                    stats.initial.main = size;
                    console.log(`${colors.blue}→ 🎯 Identified as main bundle${colors.reset}`);
                } else if (file.startsWith('scripts-')) {
                    stats.initial.runtime = size;
                    console.log(`${colors.blue}→ ⚙️ Identified as scripts/runtime bundle${colors.reset}`);
                } else if (file.startsWith('polyfills-')) {
                    stats.initial.polyfills = size;
                    console.log(`${colors.blue}→ 🔧 Identified as polyfills bundle${colors.reset}`);
                } else if (file.startsWith('chunk-') || file.includes('chunk')) {
                    stats.lazy += size;
                    console.log(`${colors.magenta}→ 📦 Identified as lazy chunk: ${file}${colors.reset}`);
                } else {
                    console.log(`${colors.yellow}→ ❓ Unknown file pattern: ${file}${colors.reset}`);
                }
            });

            stats.initial.total = stats.initial.main + stats.initial.runtime + stats.initial.polyfills;
            stats.total = stats.initial.total + stats.lazy;
            stats.totalWithAssets = stats.initial.total + stats.lazy + stats.assets;
        } catch (error) {
            console.error(`${colors.red}❌ Error calculating bundle size:${colors.reset}`, error);
            console.error(`${colors.red}Error details:${colors.reset}`, error.message);
        }

        return stats;
    }

    async benchmarkApp(appName, description) {
        console.log(`\n${colors.cyan}🚀 Benchmarking ${colors.yellow}${appName}${colors.reset}...`);

        const startTime = Date.now();

        try {
            console.log(`\n${colors.yellow}🧹 Cleaning previous build...${colors.reset}`);
            execSync(`nx reset`, { stdio: 'inherit' });
            execSync(`rm -rf dist/apps/${appName}`, { stdio: 'inherit' });

            console.log(`\n${colors.yellow}🏗️  Building production bundle...${colors.reset}`);
            execSync(`nx build ${appName} --configuration=production`, { stdio: 'inherit' });

            const buildTime = Date.now() - startTime;

            const bundleSize = this.calculateBundleSize(appName);
            this.logStats(bundleSize, { appName, buildTime });

            const metrics = {
                appName,
                description,
                buildTime,
                bundleSize,
                timestamp: new Date().toISOString(),
            };

            this.metricsHistory.push(metrics);
            this.saveMetricsHistory();

            return metrics;
        } catch (error) {
            console.error(`${colors.red}❌ Error benchmarking ${appName}:${colors.reset}`, error);
            throw error;
        }
    }

    async benchmarkApps(apps, { description }) {
        console.log(
            `${colors.cyan}🚀 Starting benchmark for ${apps.length} ${apps.length === 1 ? 'app' : 'apps'}...${colors.reset}`,
        );
        const results = [];

        for (const app of apps) {
            try {
                const metrics = await this.benchmarkApp(app, description);
                results.push(metrics);
            } catch (error) {
                console.error(`${colors.red}Failed to benchmark ${app}:${colors.reset}`, error);
            }
        }

        return results;
    }

    formatSize(sizeInBytes) {
        const kb = sizeInBytes / 1024;

        if (kb > 1024) {
            const mb = kb / 1024;
            return `${kb.toFixed(2)}KB (${mb.toFixed(2)}MB)`;
        }

        return `${kb.toFixed(2)}KB`;
    }

    logStats(stats, { appName, buildTime }) {
        console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);
        console.log(`${colors.cyan} 📊 Report for ${colors.yellow}${appName} app${colors.reset}:`);
        console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);
        console.log(`${colors.green} 🕒 Build time: ${buildTime / 1000}s${colors.reset}`);
        console.log(`${colors.green} 🎯 Main bundle: ${this.formatSize(stats.initial.main)}${colors.reset}`);
        console.log(`${colors.green} ⚙️ Runtime bundle: ${this.formatSize(stats.initial.runtime)}${colors.reset}`);
        console.log(`${colors.green} 🔧 Polyfills bundle: ${this.formatSize(stats.initial.polyfills)}${colors.reset}`);
        console.log(`${colors.yellow} 📦 Initial total: ${this.formatSize(stats.initial.total)}${colors.reset}`);
        console.log(`${colors.magenta} 📦 Lazy chunks total: ${this.formatSize(stats.lazy)}${colors.reset}`);
        console.log(`${colors.blue} 📦 Bundle total: ${this.formatSize(stats.total)}${colors.reset}`);
        console.log(`${colors.cyan} 📂 Assets total: ${this.formatSize(stats.assets)}${colors.reset}`);
        console.log(`${colors.blue} 📊 Overall total: ${this.formatSize(stats.totalWithAssets)}${colors.reset}`);
        console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);
    }

    logMetricsHistory(count) {
        const metrics = this.metricsHistory
            .toSorted((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
            .slice(0, count || this.metricsHistory.length);

        console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);
        console.log(`${colors.cyan} 📊 Metrics history (${metrics.length} records):${colors.reset}`);
        console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);

        metrics
            .forEach(metrics => {
                console.log(
                    `${colors.cyan} 🗓️ Recorded on ${format(metrics.timestamp, 'dd/MM/yyyy')} at ${format(metrics.timestamp, 'HH:mm')}${colors.reset}`,
                );
                console.log(`${colors.cyan} 📝 Description: ${metrics.description || 'N/A'}${colors.reset}`);
                console.log(`${colors.cyan} 📊 App: ${metrics.appName}${colors.reset}`);
                console.log(`${colors.green} 🕒 Build time: ${metrics.buildTime / 1000}s${colors.reset}`);
                console.log(
                    `${colors.green} 🎯 Main bundle: ${this.formatSize(metrics.bundleSize.initial.main)}${colors.reset}`,
                );
                console.log(
                    `${colors.green} ⚙️ Runtime bundle: ${this.formatSize(metrics.bundleSize.initial.runtime)}${colors.reset}`,
                );
                console.log(
                    `${colors.green} 🔧 Polyfills bundle: ${this.formatSize(metrics.bundleSize.initial.polyfills)}${colors.reset}`,
                );
                console.log(
                    `${colors.yellow} 📦 Initial total: ${this.formatSize(metrics.bundleSize.initial.total)}${colors.reset}`,
                );
                console.log(
                    `${colors.magenta} 📦 Lazy chunks total: ${this.formatSize(metrics.bundleSize.lazy)}${colors.reset}`,
                );
                console.log(`${colors.blue} 📦 Bundle total: ${this.formatSize(metrics.bundleSize.total)}${colors.reset}`);
                console.log(`${colors.cyan} 📂 Assets total: ${this.formatSize(metrics.bundleSize.assets)}${colors.reset}`);
                console.log(
                    `${colors.blue} 📊 Overall total: ${this.formatSize(metrics.bundleSize.totalWithAssets)}${colors.reset}`,
                );
                console.log(`${colors.cyan}------------------------------------------------------------------${colors.reset}`);
            });
    }
}

// Parse command line arguments
const args = process.argv.slice(2);

// Help command
if (args.includes('--help') || args.includes('-h')) {
    console.log(`
${colors.cyan}Nx Workspace Build Performance Benchmarking Tool${colors.reset}

Usage:
  node nx-benchmark.js [options] [app-names...]

Options:
  --all                 Benchmark all apps in the workspace
  --description, -d     A description of the benchmark that can be used to identify it in the history
  --help, -h            Show this help message
  --history, -H         Show the metrics history

Examples:
  node nx-benchmark.js app1 app2                         Benchmark specific apps
  node nx-benchmark.js --all                             Benchmark all apps
  node nx-benchmark.js --all --description="Baseline"    Benchmark all apps with a description
  node nx-benchmark.js --history=2                       Show the last 2 records in the metrics history
  `);
    process.exit(0);
}

const workspace = new NxBenchmark();

if (args[0]?.startsWith('--history') || args[0]?.startsWith('-H')) {
    const historyArg = args.find(arg => arg.startsWith('--history=') || arg.startsWith('-H='));
    const historyCount = historyArg ? parseInt(historyArg.split('=')[1]) : null;
    workspace.logMetricsHistory(historyCount);
    process.exit(0);
}

// Determine which apps to benchmark
let appsToTest = [];
const description = args.find(arg => arg.startsWith('--description=') || arg.startsWith('-d='))?.split('=')[1] ?? '';

if (args.includes('--all')) {
    appsToTest = workspace.getAllApps();
    if (appsToTest.length === 0) {
        console.error(`${colors.red}❌ No apps found in workspace${colors.reset}`);
        process.exit(1);
    }
} else if (args.length === 0) {
    console.error(`${colors.red}❌ Please provide app names or use --all${colors.reset}`);
    console.log(`${colors.yellow}Usage: node nx-benchmark.js [app-names...] or --all${colors.reset}`);
    process.exit(1);
} else {
    appsToTest = args.filter(arg => !arg.startsWith('-'));
}

workspace.benchmarkApps(appsToTest, { description }).catch(error => {
    console.error(`${colors.red}❌ Benchmark failed:${colors.reset}`, error);
});