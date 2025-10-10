package com.dikuclient

import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.ViewModel
import java.io.File

class ClientViewModel : ViewModel() {
    val isConnected = mutableStateOf(false)
    val terminalOutput = mutableStateOf("")
    
    private var ptyMaster: Int = -1
    
    fun connect(host: String, port: Int) {
        // Validate inputs
        if (host.isEmpty()) {
            return
        }
        if (port < 1 || port > 65535) {
            return
        }
        
        // In real implementation, this would:
        // 1. Create a PTY using JNI or Android APIs
        // 2. Call Go code: mobile.StartClient(host, port, ptyFd)
        // 3. Start reading from PTY and updating terminalOutput
        
        // For now, simulate connection
        isConnected.value = true
        terminalOutput.value = "Connected to $host:$port\n\nWelcome to DikuMUD Client!\n\n"
    }
    
    fun disconnect() {
        // In real implementation: mobile.Stop()
        isConnected.value = false
        terminalOutput.value = ""
        
        if (ptyMaster >= 0) {
            // Close PTY
            ptyMaster = -1
        }
    }
    
    fun sendInput(text: String) {
        // In real implementation: mobile.SendText(text)
        // For now, just echo the input
        terminalOutput.value += "> $text\n"
    }
}
