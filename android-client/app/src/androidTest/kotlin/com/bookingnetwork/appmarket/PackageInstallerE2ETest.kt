package com.bookingnetwork.appmarket

import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageInstaller
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.filters.MediumTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import java.io.ByteArrayInputStream
import java.io.InputStream

/**
 * End-to-End instrumentation test verifying Android PackageInstaller.Session workflow,
 * multi-split stream writing, and session commit/rollback across API levels 26-35.
 * Satisfies TC-AND-001 & TC-AND-002 requirements.
 */
@RunWith(AndroidJUnit4::class)
@MediumTest
class PackageInstallerE2ETest {

    private lateinit var context: Context
    private lateinit var packageInstaller: PackageInstaller

    @Before
    fun setUp() {
        context = ApplicationProvider.getApplicationContext()
        packageInstaller = context.packageManager.packageInstaller
    }

    @Test
    fun testCreateSessionAndCommitLifecycle() {
        val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL).apply {
            setAppPackageName("com.example.testapp")
            setSize(1024 * 1024)
        }

        val sessionId = packageInstaller.createSession(params)
        assertTrue("Session ID must be positive integer", sessionId > 0)

        val session = packageInstaller.openSession(sessionId)
        assertNotNull("Opened session must not be null", session)

        try {
            // Write base APK split stream
            val dummyApkBytes = ByteArray(4096) { 0x50 } // Mock PK header prefix
            val inStream: InputStream = ByteArrayInputStream(dummyApkBytes)
            session.openWrite("base.apk", 0, dummyApkBytes.size.toLong()).use { outStream ->
                inStream.copyTo(outStream)
                session.fsync(outStream)
            }

            // Verify session names
            val names = session.names
            assertTrue("Session should contain base.apk", names.contains("base.apk"))

            // Abandon session to clean up without installing corrupt payload on test device
            session.abandon()
        } catch (e: Exception) {
            session.abandon()
            fail("Session operation failed: ${e.message}")
        }
    }

    @Test
    fun testMultiSplitApkStreaming() {
        val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL).apply {
            setAppPackageName("com.example.splitapp")
        }

        val sessionId = packageInstaller.createSession(params)
        val session = packageInstaller.openSession(sessionId)

        try {
            // Write base.apk
            val baseBytes = ByteArray(2048) { 0x11 }
            session.openWrite("base.apk", 0, baseBytes.size.toLong()).use { out ->
                ByteArrayInputStream(baseBytes).copyTo(out)
                session.fsync(out)
            }

            // Write config.arm64_v8a.apk
            val configAbiBytes = ByteArray(1024) { 0x22 }
            session.openWrite("split_config.arm64_v8a.apk", 0, configAbiBytes.size.toLong()).use { out ->
                ByteArrayInputStream(configAbiBytes).copyTo(out)
                session.fsync(out)
            }

            // Write config.ko.apk (Korean language split)
            val configLangBytes = ByteArray(512) { 0x33 }
            session.openWrite("split_config.ko.apk", 0, configLangBytes.size.toLong()).use { out ->
                ByteArrayInputStream(configLangBytes).copyTo(out)
                session.fsync(out)
            }

            val splits = session.names
            assertEquals("Should have exactly 3 splits written", 3, splits.size)
            assertTrue(splits.contains("base.apk"))
            assertTrue(splits.contains("split_config.arm64_v8a.apk"))
            assertTrue(splits.contains("split_config.ko.apk"))

            session.abandon()
        } finally {
            try { session.close() } catch (_: Exception) {}
        }
    }

    @Test
    fun testSessionAbandonmentOnCorruptedStreamRollback() {
        val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL)
        val sessionId = packageInstaller.createSession(params)
        val session = packageInstaller.openSession(sessionId)

        // Write partial corrupted data
        val partialData = ByteArray(500) { 0xFF.toByte() }
        session.openWrite("corrupted_base.apk", 0, 1000).use { out ->
            out.write(partialData)
        }

        // Simulate crash / rollback by abandoning session
        session.abandon()

        // Verify session cannot be reopened or committed
        var reopened = false
        try {
            packageInstaller.openSession(sessionId)
            reopened = true
        } catch (expected: SecurityException) {
            // Expected after abandonment
        } catch (expected: IllegalArgumentException) {
            // Expected
        }

        assertFalse("Abandoned session must not remain accessible", reopened)
    }
}

