// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.513
package layout

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func Header() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<header class=\"bg-white border-b-gray-200 py-4 px-4 xl:px-0 duration-300 dark:bg-zinc-900 dark:border-b-zinc-800\"><div class=\"max-w-7xl mx-auto items-center justify-between md:flex\"><div class=\"flex items-center\"><a href=\"/\" class=\"font-semibold text-lg py-3 mb-1 text-primary-600 dark:text-white\" noprefetch><svg fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 260 80\" class=\"w-14\"><path d=\"M.700001 27.4C3.03333 26.7333 6.06667 26.1 9.8 25.5c3.7333-.6 7.8667-.9 12.4-.9 4.2667 0 7.8333.6 10.7 1.8 2.8667 1.1333 5.1333 2.7667 6.8 4.9 1.7333 2.0667 2.9333 4.6 3.6 7.6.7333 2.9333 1.1 6.1667 1.1 9.7V78H32.3V50.5c0-2.8-.2-5.1667-.6-7.1-.3333-2-.9333-3.6-1.8-4.8-.8-1.2667-1.9333-2.1667-3.4-2.7-1.4-.6-3.1333-.9-5.2-.9-1.5333 0-3.1333.1-4.8.3-1.6667.2-2.9.3667-3.7.5V78H.700001V27.4ZM54.2961 52.1c0-4.6.6667-8.6333 2-12.1 1.4-3.4667 3.2333-6.3333 5.5-8.6 2.2667-2.3333 4.8667-4.0667 7.8-5.2 2.9333-1.2 5.9333-1.8 9-1.8 7.2 0 12.8 2.2333 16.8 6.7 4.0667 4.4667 6.0999 11.1333 6.0999 20 0 .6667-.033 1.4333-.1 2.3 0 .8-.033 1.5333-.1 2.2H66.7961c.3333 4.2 1.8 7.4667 4.4 9.8 2.6667 2.2667 6.5 3.4 11.5 3.4 2.9333 0 5.6-.2667 8-.8 2.4667-.5333 4.4-1.1 5.8-1.7l1.6 9.9c-.6667.3333-1.6.7-2.8 1.1-1.1333.3333-2.4667.6333-4 .9-1.4667.3333-3.0667.6-4.8.8-1.7333.2-3.5.3-5.3.3-4.6 0-8.6-.6667-12-2-3.4-1.4-6.2-3.3-8.4-5.7-2.2-2.4667-3.8333-5.3333-4.9-8.6-1.0667-3.3333-1.6-6.9667-1.6-10.9Zm35.1-5.4c0-1.6667-.2333-3.2333-.7-4.7-.4667-1.5333-1.1667-2.8333-2.1-3.9-.8667-1.1333-1.9667-2-3.3-2.6-1.2667-.6667-2.8-1-4.6-1-1.8667 0-3.5.3667-4.9 1.1-1.4.6667-2.6 1.5667-3.6 2.7-.9333 1.1333-1.6667 2.4333-2.2 3.9-.5333 1.4667-.9 2.9667-1.1 4.5h22.5Zm20.0289 5.4c0-4.6.667-8.6333 2-12.1 1.4-3.4667 3.233-6.3333 5.5-8.6 2.267-2.3333 4.867-4.0667 7.8-5.2 2.933-1.2 5.933-1.8 9-1.8 7.2 0 12.8 2.2333 16.8 6.7 4.067 4.4667 6.1 11.1333 6.1 20 0 .6667-.033 1.4333-.1 2.3 0 .8-.033 1.5333-.1 2.2h-34.5c.333 4.2 1.8 7.4667 4.4 9.8 2.667 2.2667 6.5 3.4 11.5 3.4 2.933 0 5.6-.2667 8-.8 2.467-.5333 4.4-1.1 5.8-1.7l1.6 9.9c-.667.3333-1.6.7-2.8 1.1-1.133.3333-2.467.6333-4 .9-1.467.3333-3.067.6-4.8.8-1.733.2-3.5.3-5.3.3-4.6 0-8.6-.6667-12-2-3.4-1.4-6.2-3.3-8.4-5.7-2.2-2.4667-3.833-5.3333-4.9-8.6-1.067-3.3333-1.6-6.9667-1.6-10.9Zm35.1-5.4c0-1.6667-.233-3.2333-.7-4.7-.467-1.5333-1.167-2.8333-2.1-3.9-.867-1.1333-1.967-2-3.3-2.6-1.267-.6667-2.8-1-4.6-1-1.867 0-3.5.3667-4.9 1.1-1.4.6667-2.6 1.5667-3.6 2.7-.933 1.1333-1.667 2.4333-2.2 3.9-.533 1.4667-.9 2.9667-1.1 4.5h22.5Zm32.429 5c0 5.3333 1.267 9.5333 3.8 12.6 2.533 3 6.033 4.5 10.5 4.5 1.933 0 3.567-.0667 4.9-.2 1.4-.2 2.533-.4 3.4-.6V38c-1.067-.7333-2.5-1.4-4.3-2-1.733-.6667-3.6-1-5.6-1-4.4 0-7.633 1.5-9.7 4.5-2 3-3 7.0667-3 12.2Zm34.7 24.7c-2.4.7333-5.433 1.4-9.1 2-3.6.6-7.4.9-11.4.9-4.133 0-7.833-.6333-11.1-1.9-3.267-1.2667-6.067-3.0667-8.4-5.4-2.267-2.4-4.033-5.2667-5.3-8.6-1.2-3.4-1.8-7.2-1.8-11.4 0-4.1333.5-7.8667 1.5-11.2 1.067-3.4 2.6-6.3 4.6-8.7 2-2.4 4.433-4.2333 7.3-5.5 2.867-1.3333 6.167-2 9.9-2 2.533 0 4.767.3 6.7.9 1.933.6 3.6 1.2667 5 2V2.4l12.1-2.000002V76.4Zm26.47-7.1c3.2 0 5.534-.3667 7-1.1 1.467-.8 2.2-2.1333 2.2-4 0-1.7333-.8-3.1667-2.4-4.3-1.533-1.1333-4.1-2.3667-7.7-3.7-2.2-.8-4.233-1.6333-6.1-2.5-1.8-.9333-3.366-2-4.7-3.2-1.333-1.2-2.4-2.6333-3.2-4.3-.733-1.7333-1.1-3.8333-1.1-6.3 0-4.8 1.767-8.5667 5.3-11.3 3.534-2.8 8.334-4.2 14.4-4.2 3.067 0 6 .3 8.8.9 2.8.5333 4.9 1.0667 6.3 1.6l-2.2 9.8c-1.333-.6-3.033-1.1333-5.1-1.6-2.066-.5333-4.466-.8-7.2-.8-2.466 0-4.466.4333-6 1.3-1.533.8-2.3 2.0667-2.3 3.8 0 .8667.134 1.6333.4 2.3.334.6667.867 1.3 1.6 1.9.734.5333 1.7 1.1 2.9 1.7 1.2.5333 2.667 1.1 4.4 1.7 2.867 1.0667 5.3 2.1333 7.3 3.2 2 1 3.634 2.1667 4.9 3.5 1.334 1.2667 2.3 2.7333 2.9 4.4.6 1.6667.9 3.6667.9 6 0 5-1.866 8.8-5.6 11.4-3.666 2.5333-8.933 3.8-15.8 3.8-4.6 0-8.3-.4-11.1-1.2-2.8-.7333-4.766-1.3333-5.9-1.8l2.1-10.1c1.8.7333 3.934 1.4333 6.4 2.1 2.534.6667 5.4 1 8.6 1Z\" fill=\"#0070F3\"></path></svg></a></div><a href=\"/auth/login\" noprefetch class=\"p-3 -m-3 md:hidden\"><svg class=\"w-6 h-6 text-primary-600\" viewBox=\"0 0 24 24\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"><path d=\"M12 15C8.8299 15 6.01077 16.5306 4.21597 18.906C3.82968 19.4172 3.63653 19.6728 3.64285 20.0183C3.64773 20.2852 3.81533 20.6219 4.02534 20.7867C4.29716 21 4.67384 21 5.4272 21H18.5727C19.3261 21 19.7028 21 19.9746 20.7867C20.1846 20.6219 20.3522 20.2852 20.3571 20.0183C20.3634 19.6728 20.1703 19.4172 19.784 18.906C17.9892 16.5306 15.17 15 12 15Z\" stroke=\"currentColor\" stroke-width=\"1.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"></path> <path d=\"M12 12C14.4853 12 16.5 9.98528 16.5 7.5C16.5 5.01472 14.4853 3 12 3C9.51469 3 7.49997 5.01472 7.49997 7.5C7.49997 9.98528 9.51469 12 12 12Z\" stroke=\"currentColor\" stroke-width=\"1.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"></path></svg></a><div class=\"space-x-5 ml-auto items-center hidden md:flex\"><a class=\"btn btn-secondary js-ripple\" href=\"/auth/login\" noprefetch>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Var2 := `Login`
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var2)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a></div></div></header>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
