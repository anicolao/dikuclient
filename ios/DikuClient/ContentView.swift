import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = ClientViewModel()
    @State private var host: String = "aardmud.org"
    @State private var port: String = "23"
    @State private var isConnecting: Bool = false
    @State private var errorMessage: String = ""
    
    var body: some View {
        NavigationView {
            if viewModel.isConnected {
                TerminalView(viewModel: viewModel)
                    .navigationTitle("DikuClient")
                    .navigationBarTitleDisplayMode(.inline)
                    .toolbar {
                        ToolbarItem(placement: .navigationBarTrailing) {
                            Button("Disconnect") {
                                viewModel.disconnect()
                            }
                        }
                    }
            } else {
                VStack(spacing: 20) {
                    Text("DikuMUD Client")
                        .font(.largeTitle)
                        .padding(.top, 40)
                    
                    Text("Connect to your favorite MUD")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                    
                    Spacer()
                    
                    VStack(alignment: .leading, spacing: 15) {
                        Text("Server Details")
                            .font(.headline)
                        
                        TextField("Host", text: $host)
                            .textFieldStyle(RoundedBorderTextFieldStyle())
                            .autocapitalization(.none)
                            .disableAutocorrection(true)
                        
                        HStack {
                            Text("Port:")
                                .frame(width: 50, alignment: .leading)
                            TextField("Port", text: $port)
                                .textFieldStyle(RoundedBorderTextFieldStyle())
                                .keyboardType(.numberPad)
                        }
                        
                        if !errorMessage.isEmpty {
                            Text(errorMessage)
                                .foregroundColor(.red)
                                .font(.caption)
                        }
                        
                        Button(action: connect) {
                            HStack {
                                if isConnecting {
                                    ProgressView()
                                        .progressViewStyle(CircularProgressViewStyle())
                                        .padding(.trailing, 5)
                                }
                                Text(isConnecting ? "Connecting..." : "Connect")
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.blue)
                            .foregroundColor(.white)
                            .cornerRadius(10)
                        }
                        .disabled(isConnecting || host.isEmpty || port.isEmpty)
                    }
                    .padding(.horizontal)
                    
                    Spacer()
                    
                    Text("Mobile version 0.1.0")
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .padding(.bottom, 20)
                }
                .navigationTitle("DikuClient")
            }
        }
    }
    
    private func connect() {
        guard let portNum = Int(port) else {
            errorMessage = "Invalid port number"
            return
        }
        
        errorMessage = ""
        isConnecting = true
        
        DispatchQueue.global(qos: .userInitiated).async {
            let result = viewModel.connect(host: host, port: portNum)
            
            DispatchQueue.main.async {
                isConnecting = false
                if !result.isEmpty {
                    errorMessage = result
                }
            }
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
