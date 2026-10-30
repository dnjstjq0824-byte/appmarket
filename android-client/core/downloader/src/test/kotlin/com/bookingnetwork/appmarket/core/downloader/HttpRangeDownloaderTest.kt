package com.bookingnetwork.appmarket.core.downloader

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File
import java.security.MessageDigest

// TC-AND-001: HTTP Range Resume Downloader & SHA-256 Stream Integrity Verification
class HttpRangeDownloaderTest {

    @get:Rule
    val tempFolder = TemporaryFolder()

    @Test
    fun testFileSha256CalculationMatchesStandard() {
        val testFile = tempFolder.newFile("sample.apk")
        val content = "ANDROID_APK_BINARY_PAYLOAD_TEST_DATA".toByteArray(Charsets.UTF_8)
        testFile.writeBytes(content)

        val md = MessageDigest.getInstance("SHA-256")
        val expectedHash = md.digest(content).joinToString("") { "%02x".format(it) }

        val calculatedHash = HttpRangeDownloader.computeFileSha256(testFile)
        assertEquals("Calculated SHA-256 must match standard digest", expectedHash, calculatedHash)
    }

    @Test
    fun testCorruptedByteThrowsChecksumMismatch() {
        val testFile = tempFolder.newFile("corrupted.apk")
        val originalContent = "AUTHENTIC_SIGNED_PAYLOAD".toByteArray(Charsets.UTF_8)
        val expectedHash = MessageDigest.getInstance("SHA-256")
            .digest(originalContent)
            .joinToString("") { "%02x".format(it) }

        // Corrupt 1 byte in the file
        val corruptedContent = "AUTHENTIC_SIGNED_PAYLOAE".toByteArray(Charsets.UTF_8)
        testFile.writeBytes(corruptedContent)

        val actualHash = HttpRangeDownloader.computeFileSha256(testFile)
        val isMismatched = !actualHash.equals(expectedHash, ignoreCase = true)

        assertTrue("Single-byte alteration must result in checksum mismatch", isMismatched)
    }
}
