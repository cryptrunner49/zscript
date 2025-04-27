#include "libzscript.h"
#include <iostream>

int main(int argc, char** argv) {
    // Initialize the ZScript scripting environment
    ZScript_Init(argc, argv);

    if (argc > 1) {
        // Run ZScript script from a file
       ZScript_RunFile(argv[1]);
    } else {
        int exitCode;
        const char* source = "1 + 2;";
        const char* name = "<test>";

        // Interpret a ZScript script and capture the result
        char* result = ZScript_InterpretWithResult(const_cast<char*>(source), const_cast<char*>(name), &exitCode);
        if (exitCode == 0) {
            std::cout << "Last value: " << result << std::endl;
        } else {
            std::cout << "Execution failed with code " << exitCode << std::endl;
        }
        
        // Free the result string to prevent memory leaks
        free(result);
    }

    // Clean up ZScript scripting environment resources
    ZScript_Free();
    return 0;
}