" gitwing.vim — publish the current buffer on ttybus `files`
" Load from vimrc:
"   set runtimepath+=/path/to/gitwing/contrib/vim
"   runtime gitwing.vim

if exists('g:loaded_gitwing')
  finish
endif
let g:loaded_gitwing = 1

function! s:gitwing_pub() abort
  let l:p = expand('%:p')
  if l:p ==# '' || !filereadable(l:p)
    return
  endif
  let l:payload = l:p . ':' . line('.')
  let l:cmd = ['ttybus', 'pub', 'files', l:payload]
  if exists('*jobstart')
    call jobstart(l:cmd)
  elseif exists('*job_start')
    call job_start(l:cmd)
  else
    call system('ttybus pub files ' . shellescape(l:payload))
  endif
endfunction

augroup gitwing
  autocmd!
  autocmd BufEnter,FocusGained * call s:gitwing_pub()
augroup END
