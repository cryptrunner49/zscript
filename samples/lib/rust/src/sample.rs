use libc::{c_char, c_int};
use std::env;
use std::ffi::CString;
use std::ptr;

// FFI declarations for ZScript functions
#[link(name = "zscript")]
unsafe extern "C" {
    fn ZScript_Init(argc: c_int, argv: *const *mut c_char);
    fn ZScript_InterpretWithResult(
        csrc: *const c_char,
        cname: *const c_char,
        exit_code: *mut c_int,
    ) -> *mut c_char;
    fn ZScript_RunFile(cpath: *const c_char) -> c_int;
    fn ZScript_Free();
}

fn main() {
    // Convert command-line arguments to C-style argc/argv
    let args: Vec<String> = env::args().collect();
    let argc = args.len() as c_int;
    let mut argv: Vec<*mut c_char> = args
        .iter()
        .map(|arg| {
            CString::new(arg.as_str())
                .expect("Failed to convert argument to CString")
                .into_raw()
        })
        .collect();
    argv.push(ptr::null_mut()); // Null-terminate argv

    // Initialize the ZScript scripting environment
    unsafe {
        ZScript_Init(argc, argv.as_ptr());
    }

    // Run Seed script
    if args.len() > 1 {
        let path = match CString::new(args[1].as_str()) {
            Ok(path) => path,
            Err(e) => {
                eprintln!("Error: Invalid file path: {}", e);
                return;
            }
        };

        // Run ZScript script from a file
        unsafe { ZScript_RunFile(path.as_ptr()) };
    } else {
        // Run inline script
        let source = CString::new("1 + 2;").expect("Failed to create source CString");
        let name = CString::new("<test>").expect("Failed to create name CString");
        let mut exit_code: c_int = 0;

        // Interpret a ZScript script and capture the result
        let result = unsafe { ZScript_InterpretWithResult(source.as_ptr(), name.as_ptr(), &mut exit_code) };
        if exit_code == 0 {
            let result_str = unsafe { std::ffi::CStr::from_ptr(result) }
                .to_str()
                .expect("Failed to convert result to string");
            println!("Last value: {}", result_str);
        } else {
            println!("Execution failed with code {}", exit_code);
        }

        // Free the result string to prevent memory leaks
        unsafe {
            libc::free(result as *mut libc::c_void);
        }
    }

    // Clean up ZScript scripting environment resources
    unsafe {
        ZScript_Free();
    }

    // Clean up argv
    for arg in argv {
        if !arg.is_null() {
            unsafe {
                let _ = CString::from_raw(arg);
            }
        }
    }
}