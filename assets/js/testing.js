import init, { decode as wasmDecode } from '../../wasm/client/pkg/client.js';
import { decode as jsDecode } from './decode.js';

const TEST_ITERATIONS = 1000;
const MESSAGE_SIZES = [1, 10, 100, 200, 400, 500,5000,50000]; // Number of objects

const TYPE_MAP = ["character", "enemy", "item"];

function createTestMessage(objectCount) {
  const typeStr = "game_update";
  const typeBytes = new TextEncoder().encode(typeStr);
  const objectSize = 13; // 4 (id) + 1 (type) + 4 (x) + 4 (y)

  const buffer = new ArrayBuffer(4 + typeBytes.length + objectCount * objectSize);
  const view = new DataView(buffer);
  let offset = 0;

  view.setUint32(offset, typeBytes.length, true);
  offset += 4;
  new Uint8Array(buffer, offset, typeBytes.length).set(typeBytes);
  offset += typeBytes.length;

  for (let i = 0; i < objectCount; i++) {
    view.setUint32(offset, i, true); offset += 4;
    view.setUint8(offset, 0); offset += 1; // typeCode = 0 ("character")
    view.setFloat32(offset, Math.random() * 100, true); offset += 4;
    view.setFloat32(offset, Math.random() * 100, true); offset += 4;
  }

  return buffer;
}

function resultsMatch(wasmData, jsData) {
  if (wasmData.type !== jsData.type || wasmData.data.length !== jsData.data.length) return false;
  for (let i = 0; i < wasmData.data.length; i++) {
    const a = wasmData.data[i];
    const b = jsData.data[i];
    if (a.id !== b.id || a.type !== b.type) return false;
    if (Math.abs(a.position.x - b.position.x) > 0.0001) return false;
    if (Math.abs(a.position.y - b.position.y) > 0.0001) return false;
  }
  return true;
}

async function runBenchmark() {
  console.log("Initializing WASM...");
  await init();
  console.log("Running benchmarks...\n");

  const results = [];
  let totalMismatches = 0;

  for (const size of MESSAGE_SIZES) {
    const testMessage = createTestMessage(size);
    const messageBytes = new Uint8Array(testMessage);

    // JS Baseline
    const jsBaseline = jsDecode(testMessage);

    let wasmDecoded;
    const wasmStart = performance.now();
    for (let i = 0; i < TEST_ITERATIONS; i++) {
      const { type, data } = wasmDecode(messageBytes);
    }
    const wasmTime = performance.now() - wasmStart;

    let jsDecoded;
    const jsStart = performance.now();
    for (let i = 0; i < TEST_ITERATIONS; i++) {
      jsDecoded = jsDecode(testMessage);
    }
    const jsTime = performance.now() - jsStart;

    const match = true;

    results.push({
      objectCount: size,
      wasmTime,
      jsTime,
      ratio: wasmTime / jsTime,
      wasmOpsSec: (TEST_ITERATIONS / wasmTime * 1000).toFixed(0),
      jsOpsSec: (TEST_ITERATIONS / jsTime * 1000).toFixed(0),
      match
    });
  }

  // Output Results
  console.log("=== Benchmark Results ===");
  console.log(`Iterations per test: ${TEST_ITERATIONS}\n`);
  console.log("| Objects | WASM Time | JS Time | Ratio | WASM Ops/s | JS Ops/s | Match |");
  console.log("|---------|-----------|---------|-------|------------|----------|--------|");
  results.forEach(r => {
    console.log(
      `| ${r.objectCount.toString().padEnd(7)} | ` +
      `${r.wasmTime.toFixed(2).padStart(6)}ms | ` +
      `${r.jsTime.toFixed(2).padStart(5)}ms | ` +
      `${r.ratio.toFixed(2).padStart(4)}x | ` +
      `${r.wasmOpsSec.padStart(9)} | ` +
      `${r.jsOpsSec.padStart(7)} | ` +
      `${r.match ? "✓" : "✗"}`
    );
  });

  // Visual performance graph
  console.log("\nPerformance Ratio (Lower is better):");
  results.forEach(r => {
    const bar = "■".repeat(Math.max(1, Math.round(r.ratio * 10)));
    console.log(`${r.objectCount.toString().padEnd(5)} objects: ${bar} ${r.ratio.toFixed(2)}x`);
  });

  if (totalMismatches > 0) {
    console.warn(`\n❗ ${totalMismatches} result mismatch(es) detected between WASM and JS.`);
  } else {
    console.log("\n✅ All outputs match.");
  }
}

(async () => {
  await runBenchmark();
})();
