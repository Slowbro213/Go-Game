// testing.js - Comprehensive WASM vs JS Decoder Benchmark
import init, { decode as wasmDecode } from '../../wasm/decoder/pkg/decoder.js';
import { decode as jsDecode } from './decode.js';

// Configuration
const TEST_ITERATIONS = 10;
const MESSAGE_SIZES = [1, 10, 100, 1000, 5000, 50000]; // Objects per message

// Test data generator
function createTestMessage(objectCount) {
    const typeStr = "game_update";
    const typeLen = new TextEncoder().encode(typeStr).length;
    const objectSize = 21; // 4 (id) + 9 (type) + 4 (x) + 4 (y)
    
    const buffer = new ArrayBuffer(4 + typeLen + objectCount * objectSize);
    const view = new DataView(buffer);
    let offset = 0;
    
    // Write message type
    view.setUint32(offset, typeLen, true);
    offset += 4;
    new TextEncoder().encodeInto(typeStr, new Uint8Array(buffer, offset));
    offset += typeLen;
    
    // Write objects
    for (let i = 0; i < objectCount; i++) {
        view.setUint32(offset, i, true); // id
        offset += 4;
        new TextEncoder().encodeInto("character", new Uint8Array(buffer, offset));
        offset += 9;
        view.setFloat32(offset, Math.random() * 100, true); // x
        offset += 4;
        view.setFloat32(offset, Math.random() * 100, true); // y
        offset += 4;
    }
    
    return buffer;
}

// Benchmark runner
async function runBenchmark() {
    console.log("Initializing WASM...");
    await init();
    console.log("Running benchmarks...\n");
    
    const results = [];
    
    for (const size of MESSAGE_SIZES) {
        const testMessage = createTestMessage(size);
        
        // Warm up
        wasmDecode(testMessage);
        jsDecode(testMessage);
        
        // Benchmark WASM
        const wasmStart = performance.now();
        for (let i = 0; i < TEST_ITERATIONS; i++) {
            wasmDecode(testMessage);
        }
        const wasmTime = performance.now() - wasmStart;
        
        // Benchmark JS
        const jsStart = performance.now();
        for (let i = 0; i < TEST_ITERATIONS; i++) {
            jsDecode(testMessage);
        }
        const jsTime = performance.now() - jsStart;
        
        results.push({
            objectCount: size,
            wasmTime,
            jsTime,
            ratio: wasmTime / jsTime,
            wasmOpsSec: (TEST_ITERATIONS / wasmTime * 1000).toFixed(0),
            jsOpsSec: (TEST_ITERATIONS / jsTime * 1000).toFixed(0)
        });
    }
    
    // Display results
    console.log("=== Benchmark Results ===");
    console.log(`Iterations per test: ${TEST_ITERATIONS}\n`);
    
    console.log("| Objects | WASM Time | JS Time | Ratio | WASM Ops/s | JS Ops/s |");
    console.log("|---------|-----------|---------|-------|------------|----------|");
    
    results.forEach(r => {
        console.log(
            `| ${r.objectCount.toString().padEnd(7)} | ` +
            `${r.wasmTime.toFixed(2).padStart(6)}ms | ` +
            `${r.jsTime.toFixed(2).padStart(5)}ms | ` +
            `${r.ratio.toFixed(2).padStart(4)}x | ` +
            `${r.wasmOpsSec.padStart(9)} | ` +
            `${r.jsOpsSec.padStart(7)} |`
        );
    });
    
    // Visual chart
    console.log("\nPerformance Ratio (Lower is better):");
    results.forEach(r => {
        const bar = "■".repeat(Math.max(1, Math.round(r.ratio * 10)));
        console.log(`${r.objectCount.toString().padEnd(5)} objects: ${bar} ${r.ratio.toFixed(2)}x`);
    });
}

// Memory usage test
function testMemoryUsage() {
    const largeMessage = createTestMessage(10000);
    const samples = [];
    
    console.log("\nTesting memory usage with 10,000 objects...");
    
    // JS memory baseline
    const jsBefore = performance.memory?.usedJSHeapSize || 0;
    const jsResult = jsDecode(largeMessage);
    const jsAfter = performance.memory?.usedJSHeapSize || 0;
    
    // WASM memory
    const wasmBefore = performance.memory?.usedJSHeapSize || 0;
    const wasmResult = wasmDecode(largeMessage);
    const wasmAfter = performance.memory?.usedJSHeapSize || 0;
    
    console.log(`JS Memory Delta: ${formatBytes(jsAfter - jsBefore)}`);
    console.log(`WASM Memory Delta: ${formatBytes(wasmAfter - wasmBefore)}`);
    
    function formatBytes(bytes) {
        if (bytes === 0) return "0 Bytes";
        const k = 1024;
        const sizes = ["Bytes", "KB", "MB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }
}

// Run all tests
(async () => {
    await runBenchmark();
    //testMemoryUsage();
})();
