// apps/web-dev/src/workers/sha256.worker.ts
// Web Worker for high-performance chunked file hashing and SHA-256 pre-calculation

self.onmessage = async (event: MessageEvent<{ file: File; chunkSize?: number }>) => {
  const { file, chunkSize = 4 * 1024 * 1024 } = event.data;
  
  if (!file) {
    self.postMessage({ type: 'ERROR', error: 'No file provided' });
    return;
  }

  try {
    const totalBytes = file.size;
    let offset = 0;
    
    // For small-to-medium files or full buffer, use SubtleCrypto
    // For streaming chunks, compute progress
    const reader = file.stream().getReader();
    const chunks: Uint8Array[] = [];
    let receivedBytes = 0;

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      if (value) {
        chunks.push(value);
        receivedBytes += value.length;
        const progress = Math.min(100, Math.round((receivedBytes / totalBytes) * 100));
        self.postMessage({
          type: 'PROGRESS',
          progress,
          bytesProcessed: receivedBytes,
          totalBytes
        });
      }
    }

    // Merge chunks into contiguous buffer for WebCrypto digest
    const completeBuffer = new Uint8Array(receivedBytes);
    let position = 0;
    for (const chunk of chunks) {
      completeBuffer.set(chunk, position);
      position += chunk.length;
    }

    const hashBuffer = await crypto.subtle.digest('SHA-256', completeBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const sha256Hex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

    self.postMessage({
      type: 'COMPLETE',
      sha256: sha256Hex,
      fileSize: totalBytes
    });
  } catch (err: any) {
    self.postMessage({
      type: 'ERROR',
      error: err?.message || 'Failed to compute SHA-256 digest'
    });
  }
};
