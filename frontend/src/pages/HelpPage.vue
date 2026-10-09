<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
</script>

<template>
  <div>
    <PageHeader title="Ajuda" description="O essencial para manter o seu endereço no ar." />
    <div class="space-y-4 max-w-[70ch]">
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">Como funciona</h2>
        <ol class="mt-2 list-decimal space-y-1 pl-5 text-base-content/85">
          <li>Em <strong>Conexões</strong>, conecte sua conta da Cloudflare usando um token. As permissões necessárias são Zone · Zone · Read e Zone · DNS · Edit.</li>
          <li>Em <strong>Hosts</strong>, escolha seu domínio e crie um endereço, como casa.seudominio.com.</li>
          <li>Em <strong>Tokens</strong>, gere um código de acesso e siga as instruções para seu roteador, Docker ou Windows enviar atualizações.</li>
          <li>No <strong>Painel</strong>, confira quando cada dispositivo enviou sua última atualização.</li>
        </ol>
      </section>
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">O que significam os estados</h2>
        <dl class="mt-2 space-y-1 text-base-content/85">
          <div><dt class="inline font-semibold">Online:</dt> <dd class="inline">o dispositivo enviou uma atualização recentemente (até 15 min, por padrão).</dd></div>
          <div><dt class="inline font-semibold">Atenção:</dt> <dd class="inline">sem contato há mais que o esperado (até 1 h).</dd></div>
          <div><dt class="inline font-semibold">Offline:</dt> <dd class="inline">nenhuma atualização foi recebida dentro do prazo. Confira o dispositivo e a conexão com a internet.</dd></div>
          <div><dt class="inline font-semibold">Aguardando:</dt> <dd class="inline">o endereço foi criado, mas a primeira atualização ainda não chegou.</dd></div>
        </dl>
      </section>
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">Seu endereço atualiza, mas o acesso não funciona?</h2>
        <p class="mt-2 text-base-content/85">
          Algumas operadoras compartilham um IP público entre vários clientes. Isso se chama CGNAT e pode impedir o acesso à sua rede de fora,
          mesmo com o endereço atualizado. O HomeAlias usa o IP informado pelo dispositivo ou detectado na conexão;
          ele não abre portas nem remove essa limitação. Confira o IP com sua operadora e as configurações de acesso da sua rede.
        </p>
      </section>
    </div>
  </div>
</template>
