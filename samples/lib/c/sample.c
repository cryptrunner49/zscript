#include "libzscript.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char** argv) {
    // Initialize the ZScript scripting environment
    ZScript_Init(argc, argv);

    if (argc > 1) {
        // Run ZScript script from a file
        ZScript_RunFile(argv[1]);
    } else {
        int exitCode;
        char* result;

        // Interpret a ZScript script and capture the result
        result = ZScript_InterpretWithResult("1 + 2;", "<test>", &exitCode);

        if (exitCode == 0) {
            printf("Last value: %s\n", result);
        } else {
            printf("Execution failed with code %d\n", exitCode);
        }

        // Free the result string to prevent memory leaks
        free(result);
    }
    
    // Clean up ZScript scripting environment resources
    ZScript_Free();
    return 0;
}