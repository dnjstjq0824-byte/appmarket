package com.bookingnetwork.appmarket

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.bookingnetwork.appmarket.feature.home.AppFeedItem
import com.bookingnetwork.appmarket.feature.home.HomeScreen

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MaterialTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    val featured = listOf(
                        AppFeedItem("app-1", "secure-pay", "Secure Pay Wallet", "FINANCE", 4.9, true),
                        AppFeedItem("app-2", "cloud-auth", "Cloud Authenticator", "TOOLS", 4.8, true)
                    )
                    val top = listOf(
                        AppFeedItem("app-3", "crypto-pro", "Pro Crypto Tracker", "FINANCE", 4.7, true),
                        AppFeedItem("app-4", "open-notes", "Open Notes FOSS", "PRODUCTIVITY", 4.6, true)
                    )
                    HomeScreen(
                        featuredApps = featured,
                        topCharts = top,
                        onAppClick = { /* Handle navigation to AppDetailScreen */ }
                    )
                }
            }
        }
    }
}

