package com.bookingnetwork.appmarket.core.downloader

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.File
import java.io.FileOutputStream
import java.io.RandomAccessFile
import java.security.MessageDigest
import java.util.concurrent.atomic.AtomicBoolean

sealed class DownloadState {
    data class Progress(val bytesRead: Long, val totalBytes: Long, val percentage: Int) : DownloadState()
    data class Completed(val file: File, val sha256Hex: String) : DownloadState()
    data class Failed(val error: Throwable) : DownloadState()
    object Cancelled : DownloadState()
}

class ChecksumMismatchException(val expected: String, val actual: String) : 
    Exception("SHA-256 Checksum Mismatch! Expected: $expected, Actual: $actual")

class HttpRangeDownloader(
    private val client: OkHttpClient = OkHttpClient.Builder().build()
) {
    private val isCancelled = AtomicBoolean(false)

    fun cancel() {
        isCancelled.set(true)
    }

    fun downloadWithResume(
        downloadUrl: String,
        destinationFile: File,
        expectedSha256: String,
        expectedSizeBytes: Long = -1L
    ): Flow<DownloadState> = flow {
        isCancelled.set(false)
        val tempFile = File(destinationFile.parentFile, destinationFile.name + ".part")
        var existingBytes = if (tempFile.exists()) tempFile.length() else 0L

        // If existing file is already equal to or larger than expected, start clean
        if (expectedSizeBytes > 0 && existingBytes >= expectedSizeBytes) {
            tempFile.delete()
            existingBytes = 0L
        }

        val requestBuilder = Request.Builder().url(downloadUrl)
        if (existingBytes > 0) {
            requestBuilder.header("Range", "bytes=$existingBytes-")
        }

        val request = requestBuilder.build()
        val response = client.newCall(request).execute()

        if (!response.isSuccessful) {
            emit(DownloadState.Failed(Exception("HTTP Error: ${response.code} ${response.message}")))
            return@flow
        }

        val isPartial = response.code == 206
        val body = response.body ?: run {
            emit(DownloadState.Failed(Exception("Empty HTTP response body")))
            return@flow
        }

        val totalLength = if (isPartial) {
            existingBytes + body.contentLength()
        } else {
            existingBytes = 0L
            tempFile.delete()
            body.contentLength()
        }

        val output = RandomAccessFile(tempFile, "rw")
        output.seek(existingBytes)

        val buffer = ByteArray(64 * 1024)
        var bytesReadTotal = existingBytes

        try {
            body.byteStream().use { input ->
                while (true) {
                    if (isCancelled.get()) {
                        emit(DownloadState.Cancelled)
                        return@flow
                    }

                    val read = input.read(buffer)
                    if (read == -1) break

                    output.write(buffer, 0, read)
                    bytesReadTotal += read

                    val percentage = if (totalLength > 0) {
                        ((bytesReadTotal * 100) / totalLength).toInt()
                    } else 0

                    emit(DownloadState.Progress(bytesReadTotal, totalLength, percentage))
                }
            }
        } finally {
            output.close()
        }

        // Integrity Check: Calculate SHA-256 over entire downloaded file
        val computedSha256 = computeFileSha256(tempFile)
        if (!computedSha256.equals(expectedSha256.trim(), ignoreCase = true)) {
            tempFile.delete() // Strict deletion on tampering/corruption
            emit(DownloadState.Failed(ChecksumMismatchException(expectedSha256, computedSha256)))
            return@flow
        }

        // Atomic move to final destination
        if (destinationFile.exists()) {
            destinationFile.delete()
        }
        if (!tempFile.renameTo(destinationFile)) {
            // Fallback manual copy
            tempFile.copyTo(destinationFile, overwrite = true)
            tempFile.delete()
        }

        emit(DownloadState.Completed(destinationFile, computedSha256))
    }.flowOn(Dispatchers.IO)

    companion object {
        fun computeFileSha256(file: File): String {
            val md = MessageDigest.getInstance("SHA-256")
            file.inputStream().use { stream ->
                val buffer = ByteArray(64 * 1024)
                var bytesRead: Int
                while (stream.read(buffer).also { bytesRead = it } != -1) {
                    md.update(buffer, 0, bytesRead)
                }
            }
            return md.digest().joinToString("") { "%02x".format(it) }
        }
    }
}

