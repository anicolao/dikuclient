import SwiftUI

struct TerminalView: View {
    @ObservedObject var viewModel: ClientViewModel
    @State private var inputText: String = ""
    @FocusState private var isInputFocused: Bool
    
    var body: some View {
        VStack(spacing: 0) {
            // Terminal output area
            ScrollView {
                ScrollViewReader { proxy in
                    Text(viewModel.terminalOutput)
                        .font(.system(.body, design: .monospaced))
                        .foregroundColor(.green)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(8)
                        .background(Color.black)
                        .id("bottom")
                        .onChange(of: viewModel.terminalOutput) { _ in
                            withAnimation {
                                proxy.scrollTo("bottom", anchor: .bottom)
                            }
                        }
                }
            }
            .background(Color.black)
            
            // Input area
            HStack {
                TextField("Enter command...", text: $inputText)
                    .textFieldStyle(PlainTextFieldStyle())
                    .padding(8)
                    .background(Color.black)
                    .foregroundColor(.green)
                    .font(.system(.body, design: .monospaced))
                    .focused($isInputFocused)
                    .onSubmit {
                        sendCommand()
                    }
                
                Button(action: sendCommand) {
                    Image(systemName: "paperplane.fill")
                        .foregroundColor(.blue)
                        .padding(8)
                }
            }
            .background(Color(white: 0.1))
            .border(Color.gray, width: 1)
        }
        .onAppear {
            // Auto-focus input field
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                isInputFocused = true
            }
        }
    }
    
    private func sendCommand() {
        guard !inputText.isEmpty else { return }
        
        viewModel.sendInput(inputText + "\n")
        inputText = ""
    }
}

struct TerminalView_Previews: PreviewProvider {
    static var previews: some View {
        TerminalView(viewModel: ClientViewModel())
    }
}
