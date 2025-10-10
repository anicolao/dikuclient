import SwiftUI
import Combine
// import Dikuclient // This will be the gomobile-generated framework

class ClientViewModel: ObservableObject {
    @Published var isConnected: Bool = false
    @Published var terminalOutput: String = ""
    
    private var ptyMaster: Int32 = -1
    private var ptySlave: Int32 = -1
    
    func connect(host: String, port: Int) -> String {
        // Validate inputs
        // Note: In real implementation, this would call DikuclientValidateConnection
        // For now, we'll do basic validation
        if host.isEmpty {
            return "Host cannot be empty"
        }
        if port < 1 || port > 65535 {
            return "Invalid port number"
        }
        
        // Create PTY
        var master: Int32 = 0
        var slave: Int32 = 0
        
        if openpty(&master, &slave, nil, nil, nil) != 0 {
            return "Failed to create PTY"
        }
        
        ptyMaster = master
        ptySlave = slave
        
        // In real implementation, this would call:
        // let error = DikuclientStartClient(host, Int(port), Int(slave))
        // For now, we simulate success
        let error = ""
        
        if error.isEmpty {
            DispatchQueue.main.async {
                self.isConnected = true
            }
            
            // Start reading from PTY
            startReadingPTY()
            return ""
        } else {
            // Clean up on failure
            close(master)
            close(slave)
            return error
        }
    }
    
    func disconnect() {
        // In real implementation: DikuclientStop()
        
        if ptyMaster >= 0 {
            close(ptyMaster)
            ptyMaster = -1
        }
        if ptySlave >= 0 {
            close(ptySlave)
            ptySlave = -1
        }
        
        DispatchQueue.main.async {
            self.isConnected = false
        }
    }
    
    func sendInput(_ text: String) {
        guard ptyMaster >= 0 else { return }
        
        // In real implementation: DikuclientSendText(text)
        // For now, write directly to PTY
        if let data = text.data(using: .utf8) {
            data.withUnsafeBytes { (ptr: UnsafeRawBufferPointer) in
                write(ptyMaster, ptr.baseAddress, data.count)
            }
        }
    }
    
    private func startReadingPTY() {
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let self = self else { return }
            
            var buffer = [UInt8](repeating: 0, count: 4096)
            
            while self.ptyMaster >= 0 {
                let bytesRead = read(self.ptyMaster, &buffer, buffer.count)
                
                if bytesRead <= 0 {
                    break
                }
                
                if let output = String(bytes: buffer[0..<bytesRead], encoding: .utf8) {
                    DispatchQueue.main.async {
                        self.terminalOutput += output
                    }
                }
            }
        }
    }
}
